# High Availability

Ivory is known for its reliability: it is very stable and typically "just works." However, there are many things that can happen in a distributed environment like Kubernetes that can affect Ivory uptime, including:

- The database storage disk fails or some other hardware failure occurs
- The network on which the database resides becomes unreachable
- The host operating system becomes unstable and crashes
- A key database file becomes corrupted
- A data center is lost
- A Kubernetes component (e.g. a Service) is accidentally deleted

There may also be downtime events that are due to the normal case of operations, such as performing a minor upgrade, security patching of operating system, hardware upgrade, or other maintenance.

The good news: IVYO is prepared for this, and your Ivory cluster is protected from many of these scenarios. However, to maximize your high availability (HA), let's first scale up your Ivory cluster.

## HA Ivory: Adding Replicas to your Ivory Cluster

IVYO provides several ways to add replicas to make a HA cluster:

- Increase the `spec.instances.replicas` value
- Add an additional entry in `spec.instances`

For the purposes of this tutorial, we will go with the first method and set `spec.instances.replicas` to `2`. Your manifest should look similar to:

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

Apply these updates to your Ivory cluster with the following command:

```
kubectl apply -k examples/kustomize/ivory
```

Within moment, you should see a new Ivory instance initializing! You can see all of your Ivory Pods for the `hippo` cluster by running the following command:

```
kubectl -n ivory-operator get pods \
  --selector=ivory-operator.ivorysql.org/cluster=hippo,ivory-operator.ivorysql.org/instance-set
```

Let's test our high availability set up.

## Testing Your HA Cluster

An important part of building a resilient Ivory environment is testing its resiliency, so let's run a few tests to see how IVYO performs under pressure!

### Test #1: Remove a Service

Let's try removing the primary Service that our application is connected to. This test does not actually require a HA Ivory cluster, but it will demonstrate IVYO's ability to react to environmental changes and heal things to ensure your applications can stay up.

Recall in the [connecting a Ivory cluster](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/connect-cluster.md) that we observed the Services that IVYO creates, e.g.:

```
kubectl -n ivory-operator get svc \
  --selector=ivory-operator.ivorysql.org/cluster=hippo
```

yields something similar to:

```
NAME              TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
hippo-ha          ClusterIP   10.103.73.92   <none>        5432/TCP   4h8m
hippo-ha-config   ClusterIP   None           <none>        <none>     4h8m
hippo-pods        ClusterIP   None           <none>        <none>     4h8m
hippo-primary     ClusterIP   None           <none>        5432/TCP   4h8m
hippo-replicas    ClusterIP   10.98.110.215  <none>        5432/TCP   4h8m
```

We also mentioned that the application is connected to the `hippo-primary` Service. What happens if we were to delete this Service?

```
kubectl -n ivory-operator delete svc hippo-primary
```

This would seem like it could create a downtime scenario, but run the above selector again:

```
kubectl -n ivory-operator get svc \
  --selector=ivory-operator.ivorysql.org/cluster=hippo
```

You should see something similar to:

```
NAME              TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
hippo-ha          ClusterIP   10.103.73.92   <none>        5432/TCP   4h8m
hippo-ha-config   ClusterIP   None           <none>        <none>     4h8m
hippo-pods        ClusterIP   None           <none>        <none>     4h8m
hippo-primary     ClusterIP   None           <none>        5432/TCP   3s
hippo-replicas    ClusterIP   10.98.110.215  <none>        5432/TCP   4h8m
```

Wow -- IVYO detected that the primary Service was deleted and it recreated it! Based on how your application connects to Ivory, it may not have even noticed that this event took place!

Now let's try a more extreme downtime event.

### Test #2: Remove the Primary StatefulSet

[StatefulSets](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/) are a Kubernetes object that provide helpful mechanisms for managing Pods that interface with stateful applications, such as databases. They provide a stable mechanism for managing Pods to help ensure data is retrievable in a predictable way.

What happens if we remove the StatefulSet that is pointed to the Pod that represents the Ivory primary? First, let's determine which Pod is the primary. We'll store it in an environmental variable for convenience.

```
PRIMARY_POD=$(kubectl -n ivory-operator get pods \
  --selector=ivory-operator.ivorysql.org/role=master \
  -o jsonpath='{.items[*].metadata.labels.ivory-operator\.ivorysql\.org/instance}')
```

Inspect the environmental variable to see which Pod is the current primary:

```
echo $PRIMARY_POD
```

should yield something similar to:

```
hippo-instance1-zj5s
```

We can use the value above to delete the StatefulSet associated with the current Ivory primary instance:

```
kubectl delete sts -n ivory-operator "${PRIMARY_POD}"
```

Let's see what happens. Try getting all of the StatefulSets for the Ivory instances in the `hippo` cluster:

```
kubectl get sts -n ivory-operator \
  --selector=ivory-operator.ivorysql.org/cluster=hippo,ivory-operator.ivorysql.org/instance
```

You should see something similar to:

```
NAME                   READY   AGE
hippo-instance1-6kbw   1/1     15m
hippo-instance1-zj5s   0/1     1s
```

IVYO recreated the StatefulSet that was deleted! After this "catastrophic" event, IVYO proceeds to heal the Ivory instance so it can rejoin the cluster. We cover the high availability process in greater depth later in the documentation.

What about the other instance? We can see that it became the new primary though the following command:

```
kubectl -n ivory-operator get pods \
  --selector=ivory-operator.ivorysql.org/role=master \
  -o jsonpath='{.items[*].metadata.labels.ivory-operator\.ivorysql\.org/instance}'
```

which should yield something similar to:

```
hippo-instance1-6kbw
```

You can test that the failover successfully occurred in a few ways. You can connect to the example Keycloak application that we [deployed in the previous section](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/connect-cluster.md). Based on Keycloak's connection retry logic, you may need to wait a moment for it to reconnect, but you will see it connected and resume being able to read and write data. You can also connect to the Ivory instance directly and execute the following command:

```
SELECT NOT pg_catalog.pg_is_in_recovery() is_primary;
```

If it returns `true` (or `t`), then the Ivory instance is a primary!

What if IVYO was down during the downtime event? Failover would still occur: the Ivory HA system works independently of IVYO and can maintain its own uptime. IVYO will still need to assist with some of the healing aspects, but your application will still maintain read/write connectivity to your Ivory cluster!

## Synchronous Replication

IvorySQL supports synchronous replication, which is a replication mode designed to limit the risk of transaction loss. Synchronous replication waits for a transaction to be written to at least one additional server before it considers the transaction to be committed. For more information on synchronous replication, please read about IVYO's [high availability architecture](https://github.com/CrunchyData/postgres-operator/blob/master/docs/content/architecture/high-availability.md#synchronous-replication-guarding-against-transactions-loss)

To add synchronous replication to your Ivory cluster, you can add the following to your spec:

```yaml
spec:
  patroni:
    dynamicConfiguration:
      synchronous_mode: true
```

While PostgreSQL defaults [`synchronous_commit`](https://www.postgresql.org/docs/current/runtime-config-wal.html#GUC-SYNCHRONOUS-COMMIT) to `on`, you may also want to explicitly set it, in which case the above block becomes:

```yaml
spec:
  patroni:
    dynamicConfiguration:
      synchronous_mode: true
      postgresql:
        parameters:
          synchronous_commit: "on"
```

Note that Patroni, which manages many aspects of the cluster's availability, will favor availability over synchronicity. This means that if a synchronous replica goes down, Patroni will allow for asynchronous replication to continue as well as writes to the primary. However, if you want to disable all writing if there are no synchronous replicas available, you would have to enable `synchronous_mode_strict`, i.e.:

```yaml
spec:
  patroni:
    dynamicConfiguration:
      synchronous_mode: true
      synchronous_mode_strict: true
```

## Affinity

[Kubernetes affinity](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/) rules, which include Pod anti-affinity and Node affinity, can help you to define where you want your workloads to reside. Pod anti-affinity is important for high availability: when used correctly, it ensures that your Ivory instances are distributed amongst different Nodes. Node affinity can be used to assign instances to specific Nodes, e.g. to utilize hardware that's optimized for databases.

### Understanding Pod Labels

IVYO sets up several labels for Ivory cluster management that can be used for Pod anti-affinity or affinity rules in general. These include:

- `ivory-operator.ivorysql.org/cluster`: This is assigned to all managed Pods in a Ivory cluster. The value of this label is the name of your Ivory cluster, in this case: `hippo`.
- `ivory-operator.ivorysql.org/instance-set`: This is assigned to all Ivory instances within a group of `spec.instances`. In the example above, the value of this label is `instance1`. If you do not assign a label, the value is automatically set by IVYO using a `NN` format, e.g. `00`.
- `ivory-operator.ivorysql.org/instance`: This is a unique label assigned to each Ivory instance containing the name of the Ivory instance.

Let's look at how we can set up affinity rules for our Ivory cluster to help improve high availability.

### Pod Anti-affinity

Kubernetes has two types of Pod anti-affinity:

- Preferred: With preferred (`preferredDuringSchedulingIgnoredDuringExecution`) Pod anti-affinity, Kubernetes will make a best effort to schedule Pods matching the anti-affinity rules to different Nodes. However, if it is not possible to do so, then Kubernetes may schedule one or more Pods to the same Node.
- Required: With required (`requiredDuringSchedulingIgnoredDuringExecution`) Pod anti-affinity, Kubernetes mandates that each Pod matching the anti-affinity rules **must** be scheduled to different Nodes. However, a Pod may not be scheduled if Kubernetes cannot find a Node that does not contain a Pod matching the rules.

There is a trade-off with these two types of pod anti-affinity: while "required" anti-affinity will ensure that all the matching Pods are scheduled on different Nodes, if Kubernetes cannot find an available Node, your Ivory instance may not be scheduled. Likewise, while "preferred" anti-affinity will make a best effort to scheduled your Pods on different Nodes, Kubernetes may compromise and schedule more than one Ivory instance of the same cluster on the same Node.

By understanding these trade-offs, the makeup of your Kubernetes cluster, and your requirements, you can choose the method that makes the most sense for your Ivory deployment. We'll show examples of both methods below!

#### Using Preferred Pod Anti-Affinity

First, let's deploy our Ivory cluster with preferred Pod anti-affinity. Note that if you have a single-node Kubernetes cluster, you will not see your Ivory instances deployed to different nodes. However, your Ivory instances _will_ be deployed.

We can set up our HA Ivory cluster with preferred Pod anti-affinity like so:

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
      dataVolumeClaimSpec:
        accessModes:
        - "ReadWriteOnce"
        resources:
          requests:
            storage: 1Gi
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
          - weight: 1
            podAffinityTerm:
              topologyKey: kubernetes.io/hostname
              labelSelector:
                matchLabels:
                  ivory-operator.ivorysql.org/cluster: hippo
                  ivory-operator.ivorysql.org/instance-set: instance1
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

Apply those changes in your Kubernetes cluster.

Let's take a closer look at this section:

```
affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 1
      podAffinityTerm:
        topologyKey: kubernetes.io/hostname
        labelSelector:
          matchLabels:
            ivory-operator.ivorysql.org/cluster: hippo
            ivory-operator.ivorysql.org/instance-set: instance1
```

`spec.instances.affinity.podAntiAffinity` follows the standard Kubernetes [Pod anti-affinity spec](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/). The values for the `matchLabels` are derived from what we described in the previous section: `ivory-operator.ivorysql.org/cluster` is set to our cluster name of `hippo`, and `ivory-operator.ivorysql.org/instance-set` is set to the instance set name of `instance1`. We choose a `topologyKey` of `kubernetes.io/hostname`, which is standard in Kubernetes clusters.

Preferred Pod anti-affinity will perform a best effort to schedule your Ivory Pods to different nodes. Let's see how you can require your Ivory Pods to be scheduled to different nodes.

#### Using Required Pod Anti-Affinity

Required Pod anti-affinity forces Kubernetes to scheduled your Ivory Pods to different Nodes. Note that if Kubernetes is unable to schedule all Pods to different Nodes, some of your Ivory instances may become unavailable.

Using the previous example, let's indicate to Kubernetes that we want to use required Pod anti-affinity for our Ivory clusters:

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
      dataVolumeClaimSpec:
        accessModes:
        - "ReadWriteOnce"
        resources:
          requests:
            storage: 1Gi
      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - topologyKey: kubernetes.io/hostname
            labelSelector:
              matchLabels:
                ivory-operator.ivorysql.org/cluster: hippo
                ivory-operator.ivorysql.org/instance-set: instance1
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

Apply those changes in your Kubernetes cluster.

If you are in a single Node Kubernetes clusters, you will notice that not all of your Ivory instance Pods will be scheduled. This is due to the `requiredDuringSchedulingIgnoredDuringExecution` preference. However, if you have enough Nodes available, you will see the Ivory instance Pods scheduled to different Nodes:

```
kubectl get pods -n ivory-operator -o wide \
  --selector=ivory-operator.ivorysql.org/cluster=hippo,ivory-operator.ivorysql.org/instance
```

### Node Affinity

Node affinity can be used to assign your Ivory instances to Nodes with specific hardware or to guarantee a Ivory instance resides in a specific zone. Node affinity can be set within the `spec.instances.affinity.nodeAffinity` attribute, following the standard Kubernetes [node affinity spec](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/).

Let's see an example with required Node affinity. Let's say we have a set of Nodes that are reserved for database usage that have a label `workload-role=db`. We can create a Ivory cluster with a required Node affinity rule to scheduled all of the databases to those Nodes using the following configuration:

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
      dataVolumeClaimSpec:
        accessModes:
        - "ReadWriteOnce"
        resources:
          requests:
            storage: 1Gi
      affinity:
        nodeAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            nodeSelectorTerms:
            - matchExpressions:
              - key: workload-role
                operator: In
                values:
                - db
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

## Pod Topology Spread Constraints

In addition to affinity and anti-affinity settings, [Kubernetes Pod Topology Spread Constraints](https://kubernetes.io/docs/concepts/workloads/pods/pod-topology-spread-constraints/) can also help you to define where you want your workloads to reside. However, while PodAffinity allows any number of Pods to be added to a qualifying topology domain, and PodAntiAffinity allows only one Pod to be scheduled into a single topology domain, topology spread constraints allow you to distribute Pods across different topology domains with a finer level of control.

### API Field Configuration

The spread constraint [API fields](https://kubernetes.io/docs/concepts/workloads/pods/pod-topology-spread-constraints/#spread-constraints-for-pods) can be configured for instance, PgBouncer and pgBackRest repo host pods. The basic configuration is as follows:

```
      topologySpreadConstraints:
      - maxSkew: <integer>
        topologyKey: <string>
        whenUnsatisfiable: <string>
        labelSelector: <object>
```

where "maxSkew" describes the maximum degree to which Pods can be unevenly distributed, "topologyKey" is the key that defines a topology in the Nodes' Labels, "whenUnsatisfiable" specifies what action should be taken when "maxSkew" can't be satisfied, and "labelSelector" is used to find matching Pods.

### Example Spread Constraints

To help illustrate how you might use this with your cluster, we can review examples for configuring spread constraints on our Instance and pgBackRest repo host Pods. For this example, assume we have a three node Kubernetes cluster where the first node is labeled with `my-node-label=one`, the second node is labeled with `my-node-label=two` and the final node is labeled `my-node-label=three`. The label key `my-node-label` will function as our `topologyKey`. Note all three nodes in our examples will be schedulable, so a Pod could live on any of the three Nodes.

#### Instance Pod Spread Constraints

To begin, we can set our topology spread constraints on our cluster Instance Pods. Given this configuration

```
  instances:
    - name: instance1
      replicas: 5
      topologySpreadConstraints:
        - maxSkew: 1
          topologyKey: my-node-label
          whenUnsatisfiable: DoNotSchedule
          labelSelector:
            matchLabels:
              ivory-operator.ivorysql.org/instance-set: instance1
```

we will expect 5 Instance pods to be created. Each of these Pods will have the standard `ivory-operator.ivorysql.org/instance-set: instance1` Label set, so each Pod will be properly counted when determining the `maxSkew`. Since we have 3 nodes with a `maxSkew` of 1 and we've set `whenUnsatisfiable` to `DoNotSchedule`, we should see 2 Pods on 2 of the nodes and 1 Pod on the remaining Node, thus ensuring our Pods are distributed as evenly as possible.

#### pgBackRest Repo Pod Spread Constraints

We can also set topology spread constraints on our cluster's pgBackRest repo host pod. While we normally will only have a single pod per cluster, we could use a more generic label to add a preference that repo host Pods from different clusters are distributed among our Nodes. For example, by setting our `matchLabel` value to `ivory-operator.ivorysql.org/pgbackrest: ""` and our `whenUnsatisfiable` value to `ScheduleAnyway`, we will allow our repo host Pods to be scheduled no matter what Nodes may be available, but attempt to minimize skew as much as possible.

```
      repoHost:
        topologySpreadConstraints:
        - maxSkew: 1
          topologyKey: my-node-label
          whenUnsatisfiable: ScheduleAnyway
          labelSelector:
            matchLabels:
              ivory-operator.ivorysql.org/pgbackrest: ""
```

#### Putting it All Together

Now that each of our Pods has our desired Topology Spread Constraints defined, let's put together a complete cluster definition:

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
      replicas: 5
      topologySpreadConstraints:
        - maxSkew: 1
          topologyKey: my-node-label
          whenUnsatisfiable: DoNotSchedule
          labelSelector:
            matchLabels:
              ivory-operator.ivorysql.org/instance-set: instance1
      dataVolumeClaimSpec:
        accessModes:
        - "ReadWriteOnce"
        resources:
          requests:
            storage: 1G
  backups:
    pgbackrest:
      image: {{< param imagePGBackrest >}}
      repoHost:
        topologySpreadConstraints:
        - maxSkew: 1
          topologyKey: my-node-label
          whenUnsatisfiable: ScheduleAnyway
          labelSelector:
            matchLabels:
              ivory-operator.ivorysql.org/pgbackrest: ""
      repos:
      - name: repo1
        volume:
          volumeClaimSpec:
            accessModes:
            - "ReadWriteOnce"
            resources:
              requests:
                storage: 1G
```

You can then apply those changes in your Kubernetes cluster.

Once your cluster finishes deploying, you can check that your Pods are assigned to the correct Nodes:

```
kubectl get pods -n ivory-operator -o wide --selector=ivory-operator.ivorysql.org/cluster=hippo
```

## Next Steps

We've now seen how IVYO helps your application stay "always on" with your Ivory database. Now let's explore how IVYO can minimize or eliminate downtime for operations that would normally cause that, such as [resizing your Ivory cluster](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/resize-cluster.md).
