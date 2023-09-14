# Monitoring
While having [high availability](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/high-availability.md) and
[disaster recovery](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/disaster-recovery.md) systems in place helps in the
event of something going wrong with your IvorySQL cluster, monitoring helps you anticipate
problems before they happen. Additionally, monitoring can help you diagnose and resolve issues that
may cause degraded performance rather than downtime.

Let's look at how IVYO allows you to enable monitoring in your cluster.

## Adding the Exporter Sidecar

Let's look at how we can add the IvorySQL Exporter sidecar to your cluster using the
`kustomize/ivory` example in the [Postgres Operator examples] repository.

Monitoring tools are added using the `spec.monitoring` section of the custom resource. Currently,
the only monitoring tool supported is the IvorySQL Exporter configured with [pgMonitor].

In the `kustomize/ivory/ivory.yaml` file, add the following YAML to the spec:

```
monitoring:
  pgmonitor:
    exporter:
      image: {{< param imagePostgresExporter >}}
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
the same Namespace as your PostgresCluster. It should also contain the TLS key (`tls.key`) and TLS
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

Once the IvorySQL Exporter has been enabled in your cluster, follow the steps outlined in
[IVYO Monitoring] to install the monitoring stack. This will allow you to deploy a [pgMonitor]
configuration of [Prometheus], [Grafana], and [Alertmanager] monitoring tools in Kubernetes. These
tools will be set up by default to connect to the Exporter containers on your Ivory Pods.

## Configurate Monitoring
While the default Kustomize install should work in most Kubernetes environments, it may be
necessary to further customize the project according to your specific needs.

For instance, by default `fsGroup` is set to `26` for the `securityContext` defined for the
various Deployments comprising the IVYO Monitoring stack:

```yaml
securityContext:
  fsGroup: 26
```

In most Kubernetes environments this setting is needed to ensure processes within the container
have the permissions needed to write to any volumes mounted to each of the Pods comprising the IVYO
Monitoring stack.  However, when installing in an OpenShift environment (and more specifically when
using the `restricted` Security Context Constraint), the `fsGroup` setting should be removed
since OpenShift will automatically handle setting the proper `fsGroup` within the Pod's
`securityContext`.

Additionally, within this same section it may also be necessary to modify the `supplmentalGroups`
setting according to your specific storage configuration:

```yaml
securityContext:
  supplementalGroups : 65534
```

Therefore, the following files (located under `kustomize/monitoring`) should be modified and/or
patched (e.g. using additional overlays) as needed to ensure the `securityContext` is properly
defined for your Kubernetes environment:

- `deploy-alertmanager.yaml`
- `deploy-grafana.yaml`
- `deploy-prometheus.yaml`

And to modify the configuration for the various storage resources (i.e. PersistentVolumeClaims)
created by the IVYO Monitoring installer, the `kustomize/monitoring/pvcs.yaml` file can also
be modified.

Additionally, it is also possible to further customize the configuration for the various components
comprising the IVYO Monitoring stack (Grafana, Prometheus and/or AlertManager) by modifying the
following configuration resources:

- `alertmanager-config.yaml`
- `alertmanager-rules-config.yaml`
- `grafana-datasources.yaml`
- `prometheus-config.yaml`

Finally, please note that the default username and password for Grafana can be updated by
modifying the Grafana Secret in file `kustomize/monitoring/grafana-secret.yaml`.

## Install

Once the Kustomize project has been modified according to your specific needs, IVYO Monitoring can
then be installed using `kubectl` and Kustomize:

```shell
kubectl apply -k kustomize/monitoring
```

## Uninstall

And similarly, once IVYO Monitoring has been installed, it can uninstalled using `kubectl` and
Kustomize:

```shell
kubectl delete -k kustomize/monitoring
```

## Next Steps

Now that we can monitor our cluster, let's explore how [connection pooling](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/connection-pooling.md) can be enabled using IVYO and how it is helpful.

[pgMonitor]: https://github.com/CrunchyData/pgmonitor
[Grafana]: https://grafana.com/
[Prometheus]: https://prometheus.io/
[Alertmanager]: https://prometheus.io/docs/alerting/latest/alertmanager/