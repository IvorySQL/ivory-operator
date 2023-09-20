<h1 align="center">IVYO: The Ivory Operator from IvorySQL</h1>

[![Go Report Card](https://goreportcard.com/badge/github.com/ivorysql/ivory-operator)](https://goreportcard.com/report/github.com/ivorysql/ivory-operator)

# Introduction and Special Thanks
Ivory Operator is developed based on [CrunchyData Postgres Operator](https://github.com/CrunchyData/postgres-operator). A new operator to support [IvorySQL](https://www.ivorysql.org/) is needed. CrunchyData Postgres Operator does a great job, it has legible docs, clean codes, rigorous testing and active community. Therefore, for compatible with IvorySQL, we did some code changes, modify docs and correct yaml files in our first phase. Thanks for CrunchyData Postgres Operator especially.

# Production Ivory Made Easy

[IVYO](https://github.com/IvorySQL/ivory-operator), the [Ivory Operator](https://github.com/IvorySQL/ivory-operator) from [IvorySQL](https://ivorysql.org), gives you a **declarative Ivory** solution that automatically manages your [IvorySQL](https://ivorysql.org) clusters.

Designed for your GitOps workflows, it is easy to get started with Ivory on Kubernetes with IVYO. Within a few moments, you can have a production-grade Ivory cluster complete with high availability, disaster recovery, and monitoring, all over secure TLS communications. Even better, IVYO lets you easily customize your Ivory cluster to tailor it to your workload!

With conveniences like cloning Ivory clusters to using rolling updates to roll out disruptive changes with minimal downtime, IVYO is ready to support your Ivory data at every stage of your release pipeline. Built for resiliency and uptime, IVYO will keep your Ivory cluster in its desired state, so you do not need to worry about it.

IVYO is developed with many years of production experience in automating Ivory management on Kubernetes, providing a seamless cloud native Ivory solution to keep your data always available.

# Installation

We recommend following our Quickstart for how to install and get up and running with IVYO, the Ivory Operator from IvorySQL. However, if you can't wait to try it out, here are some instructions to get Ivory up and running on Kubernetes:

1. [Fork the Ivory Operator repository](https://github.com/IvorySQL/ivory-operator/fork) and clone it to your host machine. For example:

```sh
YOUR_GITHUB_UN="<your GitHub username>"
git clone --depth 1 "git@github.com:${YOUR_GITHUB_UN}/ivory-operator.git"
cd ivory-operator
```

2. Run the following commands
```sh
kubectl apply -k examples/kustomize/install/namespace
kubectl apply --server-side -k examples/kustomize/install/default
```

For more information please read [Tutorial](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/_index.md).

# Cloud Native Ivory for Kubernetes

IVYO, the Ivory Operator from IvorySQL, comes with all of the features you need for a complete cloud native Ivory experience on Kubernetes!

#### IvorySQL Cluster [Provisioning][provisioning]

[Create, Scale, & Delete IvorySQL clusters with ease][provisioning], while fully customizing your
Pods and IvorySQL configuration!

#### [High Availability][high-availability]

Safe, automated failover backed by a [distributed consensus high availability solution][high-availability].
Uses [Pod Anti-Affinity][k8s-anti-affinity] to help resiliency; you can configure how aggressive this can be!
Failed primaries automatically heal, allowing for faster recovery time.

Support for [standby IvorySQL clusters][multiple-cluster] that work both within and across [multiple Kubernetes clusters][multiple-cluster].

#### [Disaster Recovery][disaster-recovery]

[Backups][backups] and [restores][disaster-recovery] leverage the open source [pgBackRest][] utility and
[includes support for full, incremental, and differential backups as well as efficient delta restores][backups].
Set how long you to retain your backups. Works great with very large databases!

#### Security and [TLS][tls]

IVYO enforces that all connections are over [TLS][tls]. You can also [bring your own TLS infrastructure][tls] if you do not want to use the defaults provided by IVYO.

IVYO runs containers with locked-down settings and provides Ivory credentials in a secure, convenient way for connecting your applications to your data.

#### Advanced Replication Support

Choose between [asynchronous][high-availability] and synchronous replication
for workloads that are sensitive to losing transactions.

#### [Clone][clone]

[Create new clusters from your existing clusters or backups][clone] with efficient data cloning.

#### Pod Anti-Affinity, Node Affinity, Pod Tolerations

Have your IvorySQL clusters deployed to [Kubernetes Nodes][k8s-nodes] of your preference. Set your [pod anti-affinity][k8s-anti-affinity], node affinity, Pod tolerations, and more rules to customize your deployment topology!

#### [Scheduled Backups][backup-management]

Choose the type of backup (full, incremental, differential) and [how frequently you want it to occur][backup-management] on each IvorySQL cluster.

#### [Full Customizability][customize-cluster]

IVYO makes it easy to fully customize your Ivory cluster to tailor to your workload:

- Choose the resources for your Ivory cluster: [container resources and storage size][resize-cluster]. [Resize at any time][resize-cluster] with minimal disruption.
- - Use your own container image repository, including support `imagePullSecrets` and private repositories
- [Customize your IvorySQL configuration][customize-cluster]

#### [Namespaces][k8s-namespaces]

Deploy IVYO to watch Ivory clusters in all of your [namespaces][k8s-namespaces], or [restrict which namespaces][single-namespace] you want IVYO to manage Ivory clusters in!

[backups]: https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/backups.md
[backup-management]: https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/backup-management.md
[clone]: https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/disaster-recovery.md#clone-a-ivory-cluster
[customize-cluster]: https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/customize-cluster.md
[disaster-recovery]: https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/disaster-recovery.md
[high-availability]: https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/high-availability.md
[monitoring]: https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/monitoring.md
[multiple-cluster]: https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/architecture/disaster-recovery.md#standby-cluster-overview
[pool]: https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/connection-pooling.md
[provisioning]: https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/create-cluster.md
[resize-cluster]: https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/resize-cluster.md
[tls]: https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/customize-cluster.md#customize-tls


[k8s-anti-affinity]: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/#inter-pod-affinity-and-anti-affinity
[k8s-namespaces]: https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/
[k8s-nodes]: https://kubernetes.io/docs/concepts/architecture/nodes/

[pgBackRest]: https://www.pgbackrest.org
[pgBouncer]: https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/connection-pooling.md

## Included Components

IvorySQL containers deployed with the IvorySQL Operator include the following components:

- [IvorySQL](https://www.ivorysql.org)
  - [Contrib Modules](https://www.postgresql.org/docs/current/contrib.html)
  - [PL/Python + PL/Python 3](https://www.postgresql.org/docs/current/plpython.html)
  - [PL/Perl](https://www.postgresql.org/docs/current/plperl.html)
  - [PL/Tcl](https://www.postgresql.org/docs/current/pltcl.html)
  - [pgAudit](https://www.pgaudit.org/)
  - [pgAudit Analyze](https://github.com/pgaudit/pgaudit_analyze)
  - [pg_cron](https://github.com/citusdata/pg_cron)
  - [pg_partman](https://github.com/pgpartman/pg_partman)
  - [pgnodemx](https://github.com/CrunchyData/pgnodemx)
  - [set_user](https://github.com/pgaudit/set_user)
  - [TimescaleDB](https://github.com/timescale/timescaledb) (Apache-licensed community edition)
  - [wal2json](https://github.com/eulerto/wal2json)
- [pgBackRest](https://pgbackrest.org/)
- [pgBouncer](http://pgbouncer.github.io/)
- [pgAdmin 4](https://www.pgadmin.org/)
- [pgMonitor](https://github.com/CrunchyData/pgmonitor)
- [Patroni](https://patroni.readthedocs.io/) 
- [LLVM](https://llvm.org/) (for [JIT compilation](https://www.ivorysql.org/docs/current/jit.html))

In addition to the above, the geospatially enhanced IvorySQL + PostGIS container adds the following components:

- [PostGIS](http://postgis.net/)

[IvorySQL Operator Monitoring](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/architecture/monitoring.md) uses the following components:

- [pgMonitor](https://github.com/CrunchyData/pgmonitor)
- [Prometheus](https://github.com/prometheus/prometheus)
- [Grafana](https://github.com/grafana/grafana)
- [Alertmanager](https://github.com/prometheus/alertmanager)

## Supported Platforms

IVYO, the Ivory Operator from IvorySQL, is tested on the following platforms:

- Kubernetes 1.22-1.25
- OpenShift 4.8-4.11
- Rancher
- Google Kubernetes Engine (GKE), including Anthos
- Amazon EKS
- Microsoft AKS
- VMware Tanzu

This list only includes the platforms that the Ivory Operator is specifically
tested on as part of the release process: IVYO works on other Kubernetes
distributions as well.

# Documentation
For additional information regarding the design, configuration, and operation of the
IvorySQL Operator, please follow this catalog:

### Toturial
- [Get Start](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/getting-started.md)
- [Create Cluster](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/create-cluster.md)
- [Content Cluster](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/connect-cluster.md)
- [High Availability](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/high-availability.md)
- [Resize Cluster](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/resize-cluster.md)
- [Customize Cluster](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/customize-cluster.md)
- [User Management](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/user-management.md)
- [Backups](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/backups.md)
- [Backup Management](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/backup-management.md)
- [Disaster Recovery and Cloning](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/disaster-recovery.md)
- [Monitoring](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/monitoring.md)
- [Connection Pooling](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/connection-pooling.md)
- [Administrative Tasks](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/administrative-tasks.md)
- [Delete an Ivory Cluster](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/delete-cluster.md)

### Architecture
- [Overview](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/architecture/overview.md)
- [High Availability](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/high-availability.md)
- [Backups](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/architecture/backups.md)
- [Scheduling](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/architecture/scheduling.md)
- [User Management](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/architecture/user-management.md)
- [Monitoring](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/architecture/monitoring.md)
- [Disaster Recovery](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/architecture/disaster-recovery.md)
- [pgAdmin4](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/architecture/pgadmin4.md)
# Contributing to the Project

Want to contribute to the IvorySQL Operator project? Great! We've put together
a set of contributing guidelines that you can review here:

- [Contributing Guidelines](CONTRIBUTING.md)

Once you are ready to submit a Pull Request, please ensure you do the following:

1. Reviewing the [contributing guidelines](CONTRIBUTING.md) and ensure your
that you have followed the commit message format, added testing where
appropriate, documented your changes, etc.
1. Open up a pull request based upon the guidelines. If you are adding a new
feature, please open up the pull request on the `master` branch.
1. Please be as descriptive in your pull request as possible. If you are
referencing an issue, please be sure to include the issue in your pull request
