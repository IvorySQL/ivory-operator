# Create an Ivory Cluster

## Create
Creating an Ivory cluster is pretty simple. Using the example in the `examples/kustomize/ivory` directory, all we have to do is run:

```
kubectl apply -k examples/kustomize/ivory
```

and IVYO will create a simple Ivory cluster named `hippo` in the `ivory-operator` namespace. You can track the status of your Ivory cluster using `kubectl describe` on the `ivoryclusters.ivory-operator.ivorysql.org` custom resource:

```
kubectl -n ivory-operator describe ivoryclusters.ivory-operator.ivorysql.org hippo
```

and you can track the state of the Ivory Pod using the following command:

```
kubectl -n ivory-operator get pods \
  --selector=ivory-operator.ivorysql.org/cluster=hippo,ivory-operator.ivorysql.org/instance
```

### What Just Happened?

IVYO created an Ivory cluster based on the information provided to it in the Kustomize manifests located in the `examples/kustomize/ivory` directory. Let's better understand what happened by inspecting the `examples/kustomize/ivory/ivory.yaml` file:

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
```

When we ran the `kubectl apply` command earlier, what we did was create a `ivorycluster` custom resource in Kubernetes. IVYO detected that we added a new `ivorycluster` resource and started to create all the objects needed to run Ivory in Kubernetes!

What else happened? IVYO read the value from `metadata.name` to provide the Ivory cluster with the name `hippo`. Additionally, IVYO knew which containers to use for Ivory and pgBackRest by looking at the values in `spec.image` and `spec.backups.pgbackrest.image` respectively. The value in `spec.postgresVersion` is important as it will help IVYO track which major version of Ivory you are using.

IVYO knows how many Ivory instances to create through the `spec.instances` section of the manifest. While `name` is optional, we opted to give it the name `instance1`. We could have also created multiple replicas and instances during cluster initialization, but we will cover that more when we discuss how to [scale and create a HA Ivory cluster](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/high-availability.md).

A very important piece of your `ivorycluster` custom resource is the `dataVolumeClaimSpec` section. This describes the storage that your Ivory instance will use. It is modeled after the [Persistent Volume Claim](https://kubernetes.io/docs/concepts/storage/persistent-volumes/). If you do not provide a `spec.instances.dataVolumeClaimSpec.storageClassName`, then the default storage class in your Kubernetes environment is used.

As part of creating an Ivory cluster, we also specify information about our backup archive. IVYO uses [pgBackRest](https://pgbackrest.org/), an open source backup and restore tool designed to handle terabyte-scale backups. As part of initializing our cluster, we can specify where we want our backups and archives ([write-ahead logs or WAL](https://www.postgresql.org/docs/current/wal-intro.html)) stored. We will talk about this portion of the `ivorycluster` spec in greater depth in the [disaster recovery](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/backups.md) section of this tutorial, and also see how we can store backups in Amazon S3, Google GCS, and Azure Blob Storage.

## Troubleshooting

### IvorySQL / pgBackRest Pods Stuck in `Pending` Phase

The most common occurrence of this is due to PVCs not being bound. Ensure that you have set up your storage options correctly in any `volumeClaimSpec`. You can always update your settings and reapply your changes with `kubectl apply`.

Also ensure that you have enough persistent volumes available: your Kubernetes administrator may need to provision more.

If you are on OpenShift, you may need to set `spec.openshift` to `true`.


## Next Steps

We're up and running -- now let's [connect to our Ivory cluster](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/connect-cluster.md)!
