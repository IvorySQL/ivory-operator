---
title: "Resize a Ivory Cluster"
date:
draft: false
weight: 50
---

You did it -- the application is a success! Traffic is booming, so much so that you need to add more resources to your Ivory cluster. However, you're worried that any resize operation may cause downtime and create a poor experience for your end users.

This is where IVYO comes in: IVYO will help orchestrate rolling out any potentially disruptive changes to your cluster to minimize or eliminate and downtime for your application. To do so, we will assume that you have [deployed a high availability Ivory cluster]({{< relref "./high-availability.md" >}}) as described in the [previous section]({{< relref "./high-availability.md" >}}).

Let's dive in.

## Resize Memory and CPU

Memory and CPU resources are an important component for vertically scaling your Ivory cluster.
Coupled with [tweaks to your Ivory configuration file]({{< relref "./customize-cluster.md" >}}),
allocating more memory and CPU to your cluster can help it to perform better under load.

It's important for instances in the same high availability set to have the same resources.
IVYO lets you adjust CPU and memory within the `resources` sections of the `ivoryclusters.ivory-operator.ivorysql.org` custom resource. These include:

- `spec.instances.resources` section, which sets the resource values for the IvorySQL container,
  as well as any init containers in the associated pod and containers created by the `pgDataVolume` and `pgWALVolume` [data migration jobs]({{< relref "guides/data-migration.md" >}}).
- `spec.instances.sidecars.replicaCertCopy.resources` section, which sets the resources for the `replica-cert-copy` sidecar container.
- `spec.monitoring.pgmonitor.exporter.resources` section, which sets the resources for the `exporter` sidecar container.
- `spec.backups.pgbackrest.repoHost.resources` section, which sets the resources for the pgBackRest repo host container,
  as well as any init containers in the associated pod and containers created by the `pgBackRestVolume` [data migration job]({{< relref "guides/data-migration.md" >}}).
- `spec.backups.pgbackrest.sidecars.pgbackrest.resources` section, which sets the resources for the `pgbackrest` sidecar container.
- `spec.backups.pgbackrest.sidecars.pgbackrestConfig.resources` section, which sets the resources for the `pgbackrest-config` sidecar container.
- `spec.backups.pgbackrest.jobs.resources` section, which sets the resources for any pgBackRest backup job.
- `spec.backups.pgbackrest.restore.resources` section, which sets the resources for manual pgBackRest restore jobs.
- `spec.dataSource.ivorycluster.resources` section, which sets the resources for pgBackRest restore jobs created during the [cloning]({{< relref "./disaster-recovery.md" >}}) process.
- `spec.proxy.pgBouncer.resources` section, which sets the resources for the `pgbouncer` container.
- `spec.proxy.pgBouncer.sidecars.pgbouncerConfig.resources` section, which sets the resources for the `pgbouncer-config` sidecar container.

The layout of these `resources` sections should be familiar: they follow the same pattern as the standard Kubernetes structure for setting [container resources](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/). Note that these settings also allow for the configuration of [QoS classes](https://kubernetes.io/docs/tasks/configure-pod-container/quality-service-pod/).

For example, using the `spec.instances.resources` section, let's say we want to update our `hippo` Ivory cluster so that each instance has a limit of `2.0` CPUs and `4Gi` of memory. We can make the following changes to the manifest:

```
apiVersion: ivory-operator.ivorysql.org/v1beta1
kind: ivorycluster
metadata:
  name: hippo
spec:
  image: {{< param imageCrunchyPostgres >}}
  postgresVersion: {{< param postgresVersion >}}
  instances:
    - name: instance1
      replicas: 2
      resources:
        limits:
          cpu: 2.0
          memory: 4Gi
      dataVolumeClaimSpec:
        accessModes:
        - "ReadWriteOnce"
        resources:
          requests:
            storage: 1Gi
  backups:
    pgbackrest:
      image: {{< param imageCrunchyPGBackrest >}}
      repos:
      - name: repo1
        volume:
          volumeClaimSpec:
            accessModes:
            - "ReadWriteOnce"
            resources:
              requests:
                storage: 1Gi
```

In particular, we added the following to `spec.instances`:

```
resources:
  limits:
    cpu: 2.0
    memory: 4Gi
```

Apply these updates to your Ivory cluster with the following command:

```
kubectl apply -k kustomize/ivory
```

Now, let's watch how the rollout happens:

```
watch "kubectl -n ivory-operator get pods \
  --selector=ivory-operator.ivorysql.org/cluster=hippo,ivory-operator.ivorysql.org/instance \
  -o=jsonpath='{range .items[*]}{.metadata.name}{\"\t\"}{.metadata.labels.ivory-operator\.ivorysql\.org/role}{\"\t\"}{.status.phase}{\"\t\"}{.spec.containers[].resources.limits}{\"\n\"}{end}'"
```

Observe how each Pod is terminated one-at-a-time. This is part of a "rolling update". Because updating the resources of a Pod is a destructive action, IVYO first applies the CPU and memory changes to the replicas. IVYO ensures that the changes are successfully applied to a replica instance before moving on to the next replica.

Once all of the changes are applied, IVYO will perform a "controlled switchover": it will promote a replica to become a primary, and apply the changes to the final Ivory instance.

By rolling out the changes in this way, IVYO ensures there is minimal to zero disruption to your application: you are able to successfully roll out updates and your users may not even notice!

## Resize PVC

Your application is a success! Your data continues to grow, and it's becoming apparently that you need more disk. That's great: you can resize your PVC directly on your `ivoryclusters.ivory-operator.ivorysql.org` custom resource with minimal to zero downtime.

PVC resizing, also known as [volume expansion](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#expanding-persistent-volumes-claims), is a function of your storage class: it must support volume resizing. Additionally, PVCs can only be **sized up**: you cannot shrink the size of a PVC.

You can adjust PVC sizes on all of the managed storage instances in a Ivory instance that are using Kubernetes storage. These include:

- `spec.instances.dataVolumeClaimSpec.resources.requests.storage`: The Ivory data directory (aka your database).
- `spec.backups.pgbackrest.repos.volume.volumeClaimSpec.resources.requests.storage`: The pgBackRest repository when using "volume" storage

The above should be familiar: it follows the same pattern as the standard [Kubernetes PVC](https://kubernetes.io/docs/concepts/storage/persistent-volumes/) structure.

For example, let's say we want to update our `hippo` Ivory cluster so that each instance now uses a `10Gi` PVC and our backup repository uses a `20Gi` PVC. We can do so with the following markup:

```
apiVersion: ivory-operator.ivorysql.org/v1beta1
kind: ivorycluster
metadata:
  name: hippo
spec:
  image: {{< param imageCrunchyPostgres >}}
  postgresVersion: {{< param postgresVersion >}}
  instances:
    - name: instance1
      replicas: 2
      resources:
        limits:
          cpu: 2.0
          memory: 4Gi
      dataVolumeClaimSpec:
        accessModes:
        - "ReadWriteOnce"
        resources:
          requests:
            storage: 10Gi
  backups:
    pgbackrest:
      image: {{< param imageCrunchyPGBackrest >}}
      repos:
      - name: repo1
        volume:
          volumeClaimSpec:
            accessModes:
            - "ReadWriteOnce"
            resources:
              requests:
                storage: 20Gi
```

In particular, we added the following to `spec.instances`:

```
dataVolumeClaimSpec:
  resources:
    requests:
      storage: 10Gi
```

and added the following to `spec.backups.pgbackrest.repos.volume`:

```
volumeClaimSpec:
  accessModes:
  - "ReadWriteOnce"
  resources:
    requests:
      storage: 20Gi
```

Apply these updates to your Ivory cluster with the following command:

```
kubectl apply -k kustomize/ivory
```

### Resize PVCs With StorageClass That Does Not Allow Expansion

Not all Kubernetes Storage Classes allow for [volume expansion](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#expanding-persistent-volumes-claims). However, with IVYO, you can still resize your Ivory cluster data volumes even if your storage class does not allow it!

Let's go back to the previous example:

```yaml
apiVersion: ivory-operator.ivorysql.org/v1beta1
kind: ivorycluster
metadata:
  name: hippo
spec:
  image: {{< param imageCrunchyPostgres >}}
  postgresVersion: {{< param postgresVersion >}}
  instances:
    - name: instance1
      replicas: 2
      resources:
        limits:
          cpu: 2.0
          memory: 4Gi
      dataVolumeClaimSpec:
        accessModes:
        - "ReadWriteOnce"
        resources:
          requests:
            storage: 1Gi
  backups:
    pgbackrest:
      image: {{< param imageCrunchyPGBackrest >}}
      repos:
      - name: repo1
        volume:
          volumeClaimSpec:
            accessModes:
            - "ReadWriteOnce"
            resources:
              requests:
                storage: 20Gi
```

First, create a new instance that has the larger volume size. Call this instance `instance2`. The manifest would look like this:

```yaml
apiVersion: ivory-operator.ivorysql.org/v1beta1
kind: ivorycluster
metadata:
  name: hippo
spec:
  image: {{< param imageCrunchyPostgres >}}
  postgresVersion: {{< param postgresVersion >}}
  instances:
    - name: instance1
      replicas: 2
      resources:
        limits:
          cpu: 2.0
          memory: 4Gi
      dataVolumeClaimSpec:
        accessModes:
        - "ReadWriteOnce"
        resources:
          requests:
            storage: 1Gi
    - name: instance2
      replicas: 2
      resources:
        limits:
          cpu: 2.0
          memory: 4Gi
      dataVolumeClaimSpec:
        accessModes:
        - "ReadWriteOnce"
        resources:
          requests:
            storage: 10Gi
  backups:
    pgbackrest:
      image: {{< param imageCrunchyPGBackrest >}}
      repos:
      - name: repo1
        volume:
          volumeClaimSpec:
            accessModes:
            - "ReadWriteOnce"
            resources:
              requests:
                storage: 20Gi
```

Take note of the block that contains `instance2`:

```yaml
- name: instance2
  replicas: 2
  resources:
    limits:
      cpu: 2.0
      memory: 4Gi
  dataVolumeClaimSpec:
    accessModes:
    - "ReadWriteOnce"
    resources:
      requests:
        storage: 10Gi
```

This creates a second set of two Ivory instances, both of which come up as replicas, that have a larger PVC.

Once this new instance set is available and they are caught to the primary, you can then apply the following manifest:

```yaml
apiVersion: ivory-operator.ivorysql.org/v1beta1
kind: ivorycluster
metadata:
  name: hippo
spec:
  image: {{< param imageCrunchyPostgres >}}
  postgresVersion: {{< param postgresVersion >}}
  instances:
    - name: instance2
      replicas: 2
      resources:
        limits:
          cpu: 2.0
          memory: 4Gi
      dataVolumeClaimSpec:
        accessModes:
        - "ReadWriteOnce"
        resources:
          requests:
            storage: 10Gi
  backups:
    pgbackrest:
      image: {{< param imageCrunchyPGBackrest >}}
      repos:
      - name: repo1
        volume:
          volumeClaimSpec:
            accessModes:
            - "ReadWriteOnce"
            resources:
              requests:
                storage: 20Gi
```

This will promote one of the instances with the larger PVC to be the new primary and remove the instances with the smaller PVCs!

This method can also be used to shrink PVCs to use a smaller amount.

## Troubleshooting

### Ivory Pod Can't Be Scheduled

There are many reasons why a IvorySQL Pod may not be scheduled:

- **Resources are unavailable**. Ensure that you have a Kubernetes [Node](https://kubernetes.io/docs/concepts/architecture/nodes/) with enough resources to satisfy your memory or CPU Request.
- **PVC cannot be provisioned**. Ensure that you request a PVC size that is available, or that your PVC storage class is set up correctly.

### PVCs Do Not Resize

Ensure that your storage class supports PVC resizing. You can check that by inspecting the `allowVolumeExpansion` attribute:

```
kubectl get sc
```

If the storage class does not support PVC resizing, you can use the technique described above to resize PVCs using a second instance set.

## Next Steps

You've now resized your Ivory cluster, but how can you configure Ivory to take advantage of the new resources? Let's look at how we can [customize the Ivory cluster configuration]({{< relref "./customize-cluster.md" >}}).
