---
title: "Monitoring"
date:
draft: false
weight: 90
---

While having [high availability]({{< relref "tutorial/high-availability.md" >}}) and
[disaster recovery]({{< relref "tutorial/disaster-recovery.md" >}}) systems in place helps in the
event of something going wrong with your IvorySQL cluster, monitoring helps you anticipate
problems before they happen. Additionally, monitoring can help you diagnose and resolve issues that
may cause degraded performance rather than downtime.

Let's look at how IVYO allows you to enable monitoring in your cluster.

## Adding the Exporter Sidecar

Let's look at how we can add the IvorySQL IvorySQL Exporter sidecar to your cluster using the
`kustomize/ivory` example in the [Ivory Operator examples] repository.

Monitoring tools are added using the `spec.monitoring` section of the custom resource. Currently,
the only monitoring tool supported is the IvorySQL IvorySQL Exporter configured with [pgMonitor].

In the `kustomize/ivory/postgres.yaml` file, add the following YAML to the spec:

```
monitoring:
  pgmonitor:
    exporter:
      image: {{< param imageCrunchyExporter >}}
```

Save your changes and run:

```
kubectl apply -k kustomize/ivory
```

IVYO will detect the change and add the Exporter sidecar to all Ivory Pods that exist in your
cluster. IVYO will also do the work to allow the Exporter to connect to the database and gather
metrics that can be accessed using the [IVYO Monitoring] stack.

### Configuring TLS Encryption for the Exporter

IVYO allows you to configure the exporter sidecar to use TLS encryption. If you provide a custom TLS
Secret via the exporter spec:

```
  monitoring:
    pgmonitor:
      exporter:
        customTLSSecret:
          name: hippo.tls
```

Like other custom TLS Secrets that can be configured with IVYO, the Secret will need to be created in
the same Namespace as your ivorycluster. It should also contain the TLS key (`tls.key`) and TLS
certificate (`tls.crt`) needed to enable encryption.

```
data:
  tls.crt: <value>
  tls.key: <value>
```

After you configure TLS for the exporter, you will need to update your Prometheus deployment to use
TLS, and your connection to the exporter will be encrypted. Check out the [Prometheus] documentation
for more information on configuring TLS for [Prometheus].

## Accessing the Metrics

Once the IvorySQL IvorySQL Exporter has been enabled in your cluster, follow the steps outlined in
[IVYO Monitoring] to install the monitoring stack. This will allow you to deploy a [pgMonitor]
configuration of [Prometheus], [Grafana], and [Alertmanager] monitoring tools in Kubernetes. These
tools will be set up by default to connect to the Exporter containers on your Ivory Pods.

## Next Steps

Now that we can monitor our cluster, let's explore how [connection pooling]({{< relref "connection-pooling.md" >}}) can be enabled using IVYO and how it is helpful.

[pgMonitor]: https://github.com/ivorysql/pgmonitor
[Grafana]: https://grafana.com/
[Prometheus]: https://prometheus.io/
[Alertmanager]: https://prometheus.io/docs/alerting/latest/alertmanager/
[IVYO Monitoring]: {{< relref "installation/monitoring/_index.md" >}}
[Ivory Operator examples]: https://github.com/ivorysql/ivory-operator-examples/fork
