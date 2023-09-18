# Administrative Tasks

## Manually Restarting IvorySQL

There are times when you might need to manually restart IvorySQL. This can be done by adding or updating a custom annotation to the cluster's `spec.metadata.annotations` section. IVYO will notice the change and perform a [rolling restart](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/architecture/high-availability.md#rolling-update).

For example, if you have a cluster named `hippo` in the namespace `ivory-operator`, all you need to do is patch the hippo ivorycluster with the following:

```shell
kubectl patch ivorycluster/hippo -n ivory-operator --type merge \
  --patch '{"spec":{"metadata":{"annotations":{"restarted":"'"$(date)"'"}}}}'
```

Watch your hippo cluster: you will see the rolling update has been triggered and the restart has begun.

## Shutdown

You can shut down an Ivory cluster by setting the `spec.shutdown` attribute to `true`. You can do this by editing the manifest, or, in the case of the `hippo` cluster, executing a command like the below:

```
kubectl patch ivorycluster/hippo -n ivory-operator --type merge \
  --patch '{"spec":{"shutdown": true}}'
```

The effect of this is that all the Kubernetes workloads for this cluster are
scaled to 0. You can verify this with the following command:

```
kubectl get deploy,sts,cronjob --selector=ivory-operator.ivorysql.org/cluster=hippo -n ivory-operator

NAME                             READY   AGE
statefulset.apps/hippo-00-lwgx   0/0     1h

NAME                             SCHEDULE   SUSPEND   ACTIVE
cronjob.batch/hippo-repo1-full   @daily     True      0
```

To turn an Ivory cluster that is shut down back on, you can set `spec.shutdown` to `false`.

## Pausing Reconciliation and Rollout

You can pause the Ivory cluster reconciliation process by setting the
`spec.paused` attribute to `true`. You can do this by editing the manifest, or,
in the case of the `hippo` cluster, executing a command like the below:

```
kubectl patch ivorycluster/hippo -n ivory-operator --type merge \
  --patch '{"spec":{"paused": true}}'
```

Pausing a cluster will suspend any changes to the cluster’s current state until
reconciliation is resumed. This allows you to fully control when changes to
the ivorycluster spec are rolled out to the Ivory cluster. While paused,
no statuses are updated other than the "Progressing" condition.

To resume reconciliation of an Ivory cluster, you can either set `spec.paused`
to `false` or remove the setting from your manifest.

## Rotating TLS Certificates

Credentials should be invalidated and replaced (rotated) as often as possible
to minimize the risk of their misuse. Unlike passwords, every TLS certificate
has an expiration, so replacing them is inevitable.

In fact, IVYO automatically rotates the client certificates that it manages *before*
the expiration date on the certificate. A new client certificate will be generated
after 2/3rds of its working duration; so, for instance, a IVYO-created certificate
with an expiration date 12 months in the future will be replaced by IVYO around the
eight month mark. This is done so that you do not have to worry about running into
problems or interruptions of service with an expired certificate.

### Triggering a Certificate Rotation

If you want to rotate a single client certificate, you can regenerate the certificate
of an existing cluster by deleting the `tls.key` field from its certificate Secret.

Is it time to rotate your IVYO root certificate? All you need to do is delete the `ivyo-root-cacert` secret. IVYO will regenerate it and roll it out seamlessly, ensuring your apps continue communicating with the Ivory cluster without having to update any configuration or deal with any downtime.

```bash
kubectl delete secret ivyo-root-cacert
```

{{% notice note %}}
IVYO only updates secrets containing the generated root certificate. It does not touch custom certificates.
{{% /notice %}}

### Rotating Custom TLS Certificates

When you use your own TLS certificates with IVYO, you are responsible for replacing them appropriately.
Here's how.

IVYO automatically detects and loads changes to the contents of IvorySQL server
and replication Secrets without downtime. You or your certificate manager need
only replace the values in the Secret referenced by `spec.customTLSSecret`.

If instead you change `spec.customTLSSecret` to refer to a new Secret or new fields,
IVYO will perform a [rolling restart](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/architecture/high-availability.md#rolling-update).

{{% notice info %}}
When changing the IvorySQL certificate authority, make sure to update
[`customReplicationTLSSecret`](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/customize-cluster.md#customize-tls) as well.
{{% /notice %}}

## Changing the Primary

There may be times when you want to change the primary in your HA cluster. This can be done
using the `patroni.switchover` section of the ivorycluster spec. It allows
you to enable switchovers in your ivoryclusters, target a specific instance as the new
primary, and run a failover if your ivorycluster has entered a bad state.

Let's go through the process of performing a switchover!

First you need to update your spec to prepare your cluster to change the primary. Edit your spec
to have the following fields:

```yaml
spec:
  patroni:
    switchover:
      enabled: true
```

After you apply this change, IVYO will be looking for the trigger to perform a switchover in your
cluster. You will trigger the switchover by adding the `ivory-operator.ivorysql.org/trigger-switchover`
annotation to your custom resource. The best way to set this annotation is
with a timestamp, so you know when you initiated the change.

For example, for our `hippo` cluster, we can run the following command to trigger the switchover:

```shell
kubectl annotate -n ivory-operator ivorycluster hippo \
  ivory-operator.ivorysql.org/trigger-switchover="$(date)"
```

{{% notice tip %}}
If you want to perform another switchover you can re-run the annotation command and add the `--overwrite` flag:

```shell
kubectl annotate -n ivory-operator ivorycluster hippo --overwrite \
  ivory-operator.ivorysql.org/trigger-switchover="$(date)"
```
{{% /notice %}}

IVYO will detect this annotation and use the Patroni API to request a change to the current primary!

The roles on your database instance Pods will start changing as Patroni works. The new primary
will have the `master` role label, and the old primary will be updated to `replica`.

The status of the switch will be tracked using the `status.patroni.switchover` field. This will be set
to the value defined in your trigger annotation. If you use a timestamp as the annotation this is
another way to determine when the switchover was requested.

After the instance Pod labels have been updated and `status.patroni.switchover` has been set, the
primary has been changed on your cluster!

{{% notice info %}}
After changing the primary, we recommend that you disable switchovers by setting `spec.patroni.switchover.enabled`
to false or remove the field from your spec entirely. If the field is removed the corresponding
status will also be removed from the ivorycluster.
{{% /notice %}}


#### Targeting an instance

Another option you have when switching the primary is providing a target instance as the new
primary. This target instance will be used as the candidate when performing the switchover.
The `spec.patroni.switchover.targetInstance` field takes the name of the instance that you are switching to.

This name can be found in a couple different places; one is as the name of the StatefulSet and
another is on the database Pod as the `ivory-operator.ivorysql.org/instance` label. The
following commands can help you determine who is the current primary and what name to use as the
`targetInstance`:

```shell-session
$ kubectl get pods -l ivory-operator.ivorysql.org/cluster=hippo \
    -L ivory-operator.ivorysql.org/instance \
    -L ivory-operator.ivorysql.org/role -n ivory-operator

NAME                      READY   STATUS      RESTARTS   AGE     INSTANCE               ROLE
hippo-instance1-jdb5-0    3/3     Running     0          2m47s   hippo-instance1-jdb5   master
hippo-instance1-wm5p-0    3/3     Running     0          2m47s   hippo-instance1-wm5p   replica
```

In our example cluster `hippo-instance1-jdb5` is currently the primary meaning we want to target
`hippo-instance1-wm5p` in the switchover. Now that you know which instance is currently the
primary and how to find your `targetInstance`, let's update your cluster spec:

```yaml
spec:
  patroni:
    switchover:
      enabled: true
      targetInstance: hippo-instance1-wm5p
```

After applying this change you will once again need to trigger the switchover by annotating the
ivorycluster (see above commands). You can verify the switchover has completed by checking the
Pod role labels and `status.patroni.switchover`.

#### Failover

Finally, we have the option to failover when your cluster has entered an unhealthy state. The
only spec change necessary to accomplish this is updating the `spec.patroni.switchover.type`
field to the `Failover` type. One note with this is that a `targetInstance` is required when
performing a failover. Based on the example cluster above, assuming `hippo-instance1-wm5p` is still
a replica, we can update the spec:

```yaml
spec:
  patroni:
    switchover:
      enabled: true
      targetInstance: hippo-instance1-wm5p
      type: Failover
```

Apply this spec change and your ivorycluster will be prepared to perform the failover. Again
you will need to trigger the switchover by annotating the ivorycluster (see above commands)
and verify that the Pod role labels and `status.patroni.switchover` are updated accordingly.

{{% notice warning %}}
Errors encountered in the switchover process can leave your cluster in a bad
state. If you encounter issues, found in the operator logs, you can update the spec to fix the
issues and apply the change. Once the change has been applied, IVYO will attempt to perform the
switchover again.
{{% /notice %}}

## Next Steps

We've covered a lot in terms of building, maintaining, scaling, customizing, restarting, and expanding our Ivory cluster. However, there may come a time where we need to [delete our Ivory cluster](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/delete-cluster.md). How do we do that?
