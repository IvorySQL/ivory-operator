---
title: "Customize a Ivory Cluster"
date:
draft: false
weight: 60
---

Ivory is known for its ease of customization; IVYO helps you to roll out changes efficiently and without disruption. After [resizing the resources](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/resize-cluster.md) for our Ivory cluster in the previous step of this tutorial, lets see how we can tweak our Ivory configuration to optimize its usage of them.

## Custom Ivory Configuration

Part of the trick of managing multiple instances in a Ivory cluster is ensuring all of the configuration
changes are propagated to each of them. This is where IVYO helps: when you make a Ivory configuration
change for a cluster, IVYO will apply it to all of the Ivory instances.

For example, in our previous step we added CPU and memory limits of `2.0` and `4Gi` respectively. Let's tweak some of the Ivory settings to better use our new resources. We can do this in the `spec.patroni.dynamicConfiguration` section. Here is an example updated manifest that tweaks several settings:

```
apiVersion: ivory-operator.ivorysql.org/v1beta1
kind: IvoryCluster
metadata:
  name: hippo
spec:
  image: {{< param imageIvorySQL >}}
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
      image: {{< param imagePGBackrest >}}
      repos:
      - name: repo1
        volume:
          volumeClaimSpec:
            accessModes:
            - "ReadWriteOnce"
            resources:
              requests:
                storage: 1Gi
  patroni:
    dynamicConfiguration:
      postgresql:
        parameters:
          max_parallel_workers: 2
          max_worker_processes: 2
          shared_buffers: 1GB
          work_mem: 2MB
```

In particular, we added the following to `spec`:

```
patroni:
  dynamicConfiguration:
    postgresql:
      parameters:
        max_parallel_workers: 2
        max_worker_processes: 2
        shared_buffers: 1GB
        work_mem: 2MB
```

Apply these updates to your Ivory cluster with the following command:

```
kubectl apply -k examples/kustomize/ivory
```

IVYO will go and apply these settings, restarting each Ivory instance when necessary. You can verify that the changes are present using the Ivory `SHOW` command, e.g.

```
SHOW work_mem;
```

should yield something similar to:

```
 work_mem
----------
 2MB
```

## Customize TLS

All connections in IVYO use TLS to encrypt communication between components. IVYO sets up a PKI and certificate authority (CA) that allow you create verifiable endpoints. However, you may want to bring a different TLS infrastructure based upon your organizational requirements. The good news: IVYO lets you do this!

### How to Customize TLS

There are a few different TLS endpoints that can be customized for IVYO, including those of the Ivory cluster and controlling how Ivory instances authenticate with each other. Let's look at how we can customize TLS by defining

* a `spec.customTLSSecret`, used to both identify the cluster and encrypt communications; and
* a `spec.customReplicationTLSSecret`, used for replication authentication.

To customize the TLS for a Ivory cluster, you will need to create two Secrets in the Namespace of your Ivory cluster. One of these Secrets will be the `customTLSSecret` and the other will be the `customReplicationTLSSecret`. Both secrets contain a TLS key (`tls.key`), TLS certificate (`tls.crt`) and CA certificate (`ca.crt`) to use.

Note: If `spec.customTLSSecret` is provided you **must** also provide `spec.customReplicationTLSSecret` and both must contain the same `ca.crt`.

The custom TLS and custom replication TLS Secrets should contain the following fields (though see below for a workaround if you cannot control the field names of the Secret's `data`):

```
data:
  ca.crt: <value>
  tls.crt: <value>
  tls.key: <value>
```

For example, if you have files named `ca.crt`, `hippo.key`, and `hippo.crt` stored on your local machine, you could run the following command to create a Secret from those files:

```
kubectl create secret generic -n ivory-operator hippo-cluster.tls \
  --from-file=ca.crt=ca.crt \
  --from-file=tls.key=hippo.key \
  --from-file=tls.crt=hippo.crt
```

After you create the Secrets, you can specify the custom TLS Secret in your `ivorycluster.ivory-operator.ivorysql.org` custom resource. For example, if you created a `hippo-cluster.tls` Secret and a `hippo-replication.tls` Secret, you would add them to your Ivory cluster:

```
spec:
  customTLSSecret:
    name: hippo-cluster.tls
  customReplicationTLSSecret:
    name: hippo-replication.tls
```

If you're unable to control the key-value pairs in the Secret, you can create a mapping to tell
the Ivory Operator what key holds the expected value. That would look similar to this:

```
spec:
  customTLSSecret:
    name: hippo.tls
    items:
      - key: <tls.crt key in the referenced hippo.tls Secret>
        path: tls.crt
      - key: <tls.key key in the referenced hippo.tls Secret>
        path: tls.key
      - key: <ca.crt key in the referenced hippo.tls Secret>
        path: ca.crt
```

For instance, if the `hippo.tls` Secret had the `tls.crt` in a key named `hippo-tls.crt`, the
`tls.key` in a key named `hippo-tls.key`, and the `ca.crt` in a key named `hippo-ca.crt`,
then your mapping would look like:

```
spec:
  customTLSSecret:
    name: hippo.tls
    items:
      - key: hippo-tls.crt
        path: tls.crt
      - key: hippo-tls.key
        path: tls.key
      - key: hippo-ca.crt
        path: ca.crt
```

Note: Although the custom TLS and custom replication TLS Secrets share the same `ca.crt`, they do not share the same `tls.crt`:

* Your `spec.customTLSSecret` TLS certificate should have a Common Name (CN) setting that matches the primary Service name. This is the name of the cluster suffixed with `-primary`. For example, for our `hippo` cluster this would be `hippo-primary`.
* Your `spec.customReplicationTLSSecret` TLS certificate should have a Common Name (CN) setting that matches `_ivoryrepl`, which is the preset replication user.

As with the other changes, you can roll out the TLS customizations with `kubectl apply`.

## Labels

There are several ways to add your own custom Kubernetes [Labels](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/) to your Ivory cluster.

- Cluster: You can apply labels to any IVYO managed object in a cluster by editing the `spec.metadata.labels` section of the custom resource.
- Ivory: You can apply labels to a Ivory instance set and its objects by editing `spec.instances.metadata.labels`.
- pgBackRest: You can apply labels to pgBackRest and its objects by editing `ivoryclusters.spec.backups.pgbackrest.metadata.labels`.

## Annotations

There are several ways to add your own custom Kubernetes [Annotations](https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/) to your Ivory cluster.

- Cluster: You can apply annotations to any IVYO managed object in a cluster by editing the `spec.metadata.annotations` section of the custom resource.
- Ivory: You can apply annotations to a Ivory instance set and its objects by editing `spec.instances.metadata.annotations`.
- pgBackRest: You can apply annotations to pgBackRest and its objects by editing `spec.backups.pgbackrest.metadata.annotations`.

## Pod Priority Classes

IVYO allows you to use [pod priority classes](https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/) to indicate the relative importance of a pod by setting a `priorityClassName` field on your Ivory cluster. This can be done as follows:

- Instances: Priority is defined per instance set and is applied to all Pods in that instance set by editing the `spec.instances.priorityClassName` section of the custom resource.
- Dedicated Repo Host: Priority defined under the repoHost section of the spec is applied to the dedicated repo host by editing the `spec.backups.pgbackrest.repoHost.priorityClassName` section of the custom resource.
- Backup (manual and scheduled): Priority is defined under the `spec.backups.pgbackrest.jobs.priorityClassName` section and applies that priority to all pgBackRest backup Jobs (manual and scheduled).
- Restore (data source or in-place): Priority is defined for either a "data source" restore or an in-place restore by editing the `spec.dataSource.ivorycluster.priorityClassName` section of the custom resource.
- Data Migration: The priority defined for the first instance set in the spec (array position 0) is used for the PGDATA and WAL migration Jobs. The pgBackRest repo migration Job will use the priority class applied to the repoHost.

## Separate WAL PVCs

IvorySQL commits transactions by storing changes in its [Write-Ahead Log (WAL)](https://www.postgresql.org/docs/current/wal-intro.html). Because the way WAL files are accessed and
utilized often differs from that of data files, and in high-performance situations, it can desirable to put WAL files on separate storage volume. With IVYO, this can be done by adding
the `walVolumeClaimSpec` block to your desired instance in your ivorycluster spec, either when your cluster is created or anytime thereafter:

```
spec:
  instances:
    - name: instance
      walVolumeClaimSpec:
        accessModes:
        - "ReadWriteOnce"
        resources:
          requests:
            storage: 1Gi
```

This volume can be removed later by removing the `walVolumeClaimSpec` section from the instance. Note that when changing the WAL directory, care is taken so as not to lose any WAL files. IVYO only
deletes the PVC once there are no longer any WAL files on the previously configured volume.

## Database Initialization SQL

IVYO can run SQL for you as part of the cluster creation and initialization process. IVYO runs the SQL using the psql client so you can use meta-commands to connect to different databases, change error handling, or set and use variables. Its capabilities are described in the [psql documentation](https://www.postgresql.org/docs/current/app-psql.html).

### Initialization SQL ConfigMap

The Ivory cluster spec accepts a reference to a ConfigMap containing your init SQL file. Update your cluster spec to include the ConfigMap name, `spec.databaseInitSQL.name`, and the data key, `spec.databaseInitSQL.key`, for your SQL file. For example, if you create your ConfigMap with the following command:

```
kubectl -n ivory-operator create configmap hippo-init-sql --from-file=init.sql=/path/to/init.sql
```

You would add the following section to your ivorycluster spec:

```
spec:
  databaseInitSQL:
    key: init.sql
    name: hippo-init-sql
```

{{% notice note %}}
The ConfigMap must exist in the same namespace as your Ivory cluster.
{{% /notice %}}

After you add the ConfigMap reference to your spec, apply the change with `kubectl apply -k examples/kustomize/ivory`. IVYO will create your `hippo` cluster and run your initialization SQL once the cluster has started. You can verify that your SQL has been run by checking the `databaseInitSQL` status on your Ivory cluster. While the status is set, your init SQL will not be run again. You can check cluster status with the `kubectl describe` command:

```
kubectl -n ivory-operator describe ivoryclusters.ivory-operator.ivorysql.org hippo
```

{{% notice warning %}}

In some cases, due to how Kubernetes treats ivorycluster status, IVYO may run your SQL commands more than once. Please ensure that the commands defined in your init SQL are idempotent.

{{% /notice %}}

Now that `databaseInitSQL` is defined in your cluster status, verify database objects have been created as expected. After verifying, we recommend removing the `spec.databaseInitSQL` field from your spec. Removing the field from the spec will also remove `databaseInitSQL` from the cluster status.

### PSQL Usage
IVYO uses the psql interactive terminal to execute SQL statements in your database. Statements are passed in using standard input and the filename flag (e.g. `psql -f -`).

SQL statements are executed as superuser in the default maintenance database. This means you have full control to create database objects, extensions, or run any SQL statements that you might need.

#### Integration with User and Database Management

If you are creating users or databases, please see the [User/Database Management](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/user-management.md) documentation. Databases created through the user management section of the spec can be referenced in your initialization sql. For example, if a database `zoo` is defined:

```
spec:
  users:
    - name: hippo
      databases:
       - "zoo"
```

You can connect to `zoo` by adding the following `psql` meta-command to your SQL:

```
\c zoo
create table t_zoo as select s, md5(random()::text) from generate_Series(1,5) s;
```

#### Transaction support

By default, `psql` commits each SQL command as it completes. To combine multiple commands into a single [transaction](https://www.postgresql.org/docs/current/tutorial-transactions.html), use the [`BEGIN`](https://www.postgresql.org/docs/current/sql-begin.html) and [`COMMIT`](https://www.postgresql.org/docs/current/sql-commit.html) commands.

```
BEGIN;
create table t_random as select s, md5(random()::text) from generate_Series(1,5) s;
COMMIT;
```

#### PSQL Exit Code and Database Init SQL Status

The exit code from `psql` will determine when the `databaseInitSQL` status is set. When `psql` returns `0` the status will be set and SQL will not be run again. When `psql` returns with an error exit code the status will not be set. IVYO will continue attempting to execute the SQL as part of its reconcile loop until `psql` returns normally. If `psql` exits with a failure, you will need to edit the file in your ConfigMap to ensure your SQL statements will lead to a successful `psql` return. The easiest way to make live changes to your ConfigMap is to use the following `kubectl edit` command:

```
kubectl -n <cluster-namespace> edit configmap hippo-init-sql
```

Be sure to transfer any changes back over to your local file. Another option is to make changes in your local file and use `kubectl --dry-run` to create a template and pipe the output into `kubectl apply`:

```
kubectl create configmap hippo-init-sql --from-file=init.sql=/path/to/init.sql --dry-run=client -o yaml | kubectl apply -f -
```

{{% notice tip %}}
If you edit your ConfigMap and your changes aren't showing up, you may be waiting for IVYO to reconcile your cluster. After some time, IVYO will automatically reconcile the cluster or you can trigger reconciliation by applying any change to your cluster (e.g. with `kubectl apply -k examples/kustomize/ivory`).
{{% /notice %}}

To ensure that `psql` returns a failure exit code when your SQL commands fail, set the `ON_ERROR_STOP` [variable](https://www.postgresql.org/docs/current/app-psql.html#APP-PSQL-VARIABLES) as part of your SQL file:

```
\set ON_ERROR_STOP
\echo Any error will lead to exit code 3
create table t_random as select s, md5(random()::text) from generate_Series(1,5) s;
```

## Troubleshooting

### Changes Not Applied

If your Ivory configuration settings are not present, ensure that you are using the syntax that Ivory expects.
You can see this in the [Ivory configuration documentation](https://www.postgresql.org/docs/current/runtime-config.html).

## Next Steps

You've now seen how you can further customize your Ivory cluster, but what about [managing users and databases](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/user-management.md)? That's a great question that is answered in the [next section](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/user-management.md).
