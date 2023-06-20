---
title: "Apply Software Updates"
date:
draft: false
weight: 70
---

Did you know that Ivory releases bug fixes [once every three months](https://www.postgresql.org/developer/roadmap/)? Additionally, we periodically refresh the container images to ensure the base images have the latest software that may fix some CVEs.

It's generally good practice to keep your software up-to-date for stability and security purposes, so let's learn how IVYO helps to you accept low risk, "patch" type updates.

The good news: you do not need to update IVYO itself to apply component updates: you can update each Ivory cluster whenever you want to apply the update! This lets you choose when you want to apply updates to each of your Ivory clusters, so you can update it on your own schedule. If you have a [high availability Ivory]({{< relref "./high-availability.md" >}}) cluster, IVYO uses a rolling update to minimize or eliminate any downtime for your application.

## Applying Minor Ivory Updates

The Ivory image is referenced using the `spec.image` and looks similar to the below:

```
spec:
  image: registry.developers.ivorysql.org/crunchydata/IvorySQL-ivory:ubi8-14.2-0
```

Diving into the tag a bit further, you will notice the `14.2-0` portion. This represents the Ivory minor version (`14.2`) and the patch number of the release `0`. If the patch number is incremented (e.g. `14.2-1`), this means that the container is rebuilt, but there are no changes to the Ivory version. If the minor version is incremented (e.g. `14.2-0`), this means that there is a newer bug fix release of Ivory within the container.

To update the image, you just need to modify the `spec.image` field with the new image reference, e.g.

```
spec:
  image: registry.developers.ivorysql.org/crunchydata/IvorySQL-ivory:ubi8-14.2-1
```

You can apply the changes using `kubectl apply`. Similar to the rolling update example when we [resized the cluster]({{< relref "./resize-cluster.md" >}}), the update is first applied to the Ivory replicas, then a controlled switchover occurs, and the final instance is updated.

For the `hippo` cluster, you can see the status of the rollout by running the command below:

```
kubectl -n ivory-operator get pods \
  --selector=ivory-operator.ivorysql.org/cluster=hippo,ivory-operator.ivorysql.org/instance \
  -o=jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.metadata.labels.ivory-operator\.ivorysql\.org/role}{"\t"}{.status.phase}{"\t"}{.spec.containers[].image}{"\n"}{end}'
```

or by running a watch:

```
watch "kubectl -n ivory-operator get pods \
  --selector=ivory-operator.ivorysql.org/cluster=hippo,ivory-operator.ivorysql.org/instance \
  -o=jsonpath='{range .items[*]}{.metadata.name}{\"\t\"}{.metadata.labels.ivory-operator\.ivorysql\.org/role}{\"\t\"}{.status.phase}{\"\t\"}{.spec.containers[].image}{\"\n\"}{end}'"
```

## Rolling Back Minor Ivory Updates

This methodology also allows you to rollback changes from minor Ivory updates. You can change the `spec.image` field to your desired container image. IVYO will then ensure each Ivory instance in the cluster rolls back to the desired image.

## Applying Other Component Updates

There are other components that go into a IVYO Ivory cluster. These include pgBackRest, PgBouncer and others. Each one of these components has its own image: for example, you can find a reference to the pgBackRest image in the `spec.backups.pgbackrest.image` attribute.

Applying software updates for the other components in a Ivory cluster works similarly to the above. As pgBackRest and PgBouncer are Kubernetes [Deployments](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/), Kubernetes will help manage the rolling update to minimize disruption.

## Next Steps

Now that we know how to update our software components, let's look at how IVYO handles [disaster recovery]({{< relref "./backups.md" >}})!
