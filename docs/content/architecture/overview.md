# Overview

The goal of IVYO, the Ivory Operator from Highgo is to provide a means to quickly get
your applications up and running on Ivory for both development and
production environments. To understand how IVYO does this, we
want to give you a tour of its architecture, with explains both the architecture
of the IvorySQL Operator itself as well as recommended deployment models for
IvorySQL in production!

# IVYO Architecture

The Highgo IvorySQL Operator extends Kubernetes to provide a higher-level
abstraction for rapid creation and management of IvorySQL clusters.  The
Highgo IvorySQL Operator leverages a Kubernetes concept referred to as
"[Custom Resources](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)”
to create several
[custom resource definitions (CRDs)](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/#customresourcedefinitions)
that allow for the management of IvorySQL clusters.

The main custom resource definition is ivoryclusters.ivory-operator.ivorysql.org. This allows you to control all the information about an Ivory cluster, including:

- General information
- Resource allocation
- High availability
- Backup management
- Where and how it is deployed (affinity, tolerations, topology spread constraints)
- Disaster Recovery / standby clusters
- Monitoring

and more.

IVYO itself runs as a Deployment and is composed of a single container.

- `operator` (image: ivory-operator) - This is the heart of the IvorySQL
Operator. It contains a series of Kubernetes
[controllers](https://kubernetes.io/docs/concepts/architecture/controller/) that
place watch events on a series of native Kubernetes resources (Jobs, Pods) as
well as the Custom Resources that come with the IvorySQL Operator (Pgcluster,
Pgtask)

The main purpose of IVYO is to create and update information
around the structure of an Ivory Cluster, and to relay information about the
overall status and health of an IvorySQL cluster. The goal is to also simplify
this process as much as possible for users.

The Ivory Operator handles setting up all of the various StatefulSets, Deployments, Services and other Kubernetes objects.

You will also notice that **high-availability is enabled by default** if you deploy at least one Ivory replica. The
Highgo IvorySQL Operator uses a distributed-consensus method for IvorySQL
cluster high-availability, and as such delegates the management of each
cluster's availability to the clusters themselves. This removes the IvorySQL
Operator from being a single-point-of-failure, and has benefits such as faster
recovery times for each IvorySQL cluster. For a detailed discussion on
high-availability, please see the [High-Availability](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/architecture/high-availability.md)
section.

## Kubernetes StatefulSets: The IVYO Deployment Model

IVYO, the Ivory Operator from Highgo, uses [Kubernetes StatefulSets](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/)
for running Ivory instances, and will use [Deployments](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/) for more ephemeral services.

IVYO deploys Kubernetes Statefulsets in a way to allow for creating both different Ivory instance groups and be able to support advanced operations such as rolling updates that minimize or eliminate Ivory downtime. Additional components in our
IvorySQL cluster, such as the pgBackRest repository or an optional PgBouncer,
are deployed with Kubernetes Deployments.

With the IVYO architecture, we can also leverage Statefulsets to apply affinity and toleration rules across every Ivory instance or individual ones. For instance, we may want to force one or more of our IvorySQL replicas to run on Nodes in a different region than
our primary IvorySQL instances.

What's great about this is that IVYO manages this for you so you don't have to worry! Being aware of
this model can help you understand how the Ivory Operator gives you maximum
flexibility for your IvorySQL clusters while giving you the tools to
troubleshoot issues in production.

The last piece of this model is the use of [Kubernetes Services](https://kubernetes.io/docs/concepts/services-networking/service/)
for accessing your IvorySQL clusters and their various components. The
IvorySQL Operator puts services in front of each Deployment to ensure you have
a known, consistent means of accessing your IvorySQL components.

Note that in some production environments, there can be delays in accessing
Services during transition events. The IvorySQL Operator attempts to mitigate
delays during critical operations (e.g. failover, restore, etc.) by directly
accessing the Kubernetes Pods to perform given actions.

# Additional Architecture Information

There is certainly a lot to unpack in the overall architecture of IVYO. Understanding the architecture will help you to plan
the deployment model that is best for your environment. For more information on
the architectures of various components of the IvorySQL Operator, please read
onward!
