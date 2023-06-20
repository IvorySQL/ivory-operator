---
title: "Connect to a Ivory Cluster"
date:
draft: false
weight: 30
---

It's one thing to [create a Ivory cluster]({{< relref "./create-cluster.md" >}}); it's another thing to connect to it. Let's explore how IVYO makes it possible to connect to a Ivory cluster!

## Background: Services, Secrets, and TLS

IVYO creates a series of Kubernetes [Services](https://kubernetes.io/docs/concepts/services-networking/service/) to provide stable endpoints for connecting to your Ivory databases. These endpoints make it easy to provide a consistent way for your application to maintain connectivity to your data. To inspect what services are available, you can run the following command:

```
kubectl -n ivory-operator get svc --selector=ivory-operator.ivorysql.org/cluster=hippo
```

will yield something similar to:

```
NAME              TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
hippo-ha          ClusterIP   10.103.73.92   <none>        5432/TCP   3h14m
hippo-ha-config   ClusterIP   None           <none>        <none>     3h14m
hippo-pods        ClusterIP   None           <none>        <none>     3h14m
hippo-primary     ClusterIP   None           <none>        5432/TCP   3h14m
hippo-replicas    ClusterIP   10.98.110.215  <none>        5432/TCP   3h14m
```

You do not need to worry about most of these Services, as they are used to help manage the overall health of your Ivory cluster. For the purposes of connecting to your database, the Service of interest is called `hippo-primary`. Thanks to IVYO, you do not need to even worry about that, as that information is captured within a Secret!

When your Ivory cluster is initialized, IVYO will bootstrap a database and Ivory user that your application can access. This information is stored in a Secret named with the pattern `<clusterName>-pguser-<userName>`. For our `hippo` cluster, this Secret is called `hippo-pguser-hippo`. This Secret contains the information you need to connect your application to your Ivory database:

- `user`: The name of the user account.
- `password`: The password for the user account.
- `dbname`: The name of the database that the user has access to by default.
- `host`: The name of the host of the database.
  This references the [Service](https://kubernetes.io/docs/concepts/services-networking/service/) of the primary Ivory instance.
- `port`: The port that the database is listening on.
- `uri`: A [PostgresSQL connection URI](https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING)
  that provides all the information for logging into the Ivory database.
- `jdbc-uri`: A [PostgresSQL JDBC connection URI](https://jdbc.postgresql.org/documentation/use/) that provides
  all the information for logging into the Ivory database via the JDBC driver.

All connections are over TLS. IVYO provides its own certificate authority (CA) to allow you to securely connect your applications to your Ivory clusters. This allows you to use the [`verify-full` "SSL mode"](https://www.postgresql.org/docs/current/libpq-ssl.html#LIBPQ-SSL-SSLMODE-STATEMENTS) of Ivory, which provides eavesdropping protection and prevents MITM attacks. You can also choose to bring your own CA, which is described later in this tutorial in the [Customize Cluster]({{< relref "./customize-cluster.md" >}}) section.

### Modifying Service Type, NodePort Value and Metadata

By default, IVYO deploys Services with the `ClusterIP` Service type. Based on how you want to expose your database,
you may want to modify the Services to use a different
[Service type](https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types)
and [NodePort value](https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport).

You can modify the Services that IVYO manages from the following attributes:

- `spec.service` - this manages the Service for connecting to a Ivory primary.
- `spec.proxy.pgBouncer.service` - this manages the Service for connecting to the PgBouncer connection pooler.
- `spec.userInterface.pgAdmin.service` - this manages the Service for connecting to the pgAdmin management tool.

For example, say you want to set the Ivory primary to use a `NodePort` service, a specific `nodePort` value, and set
a specific annotation and label, you would add the following to your manifest:

```yaml
spec:
  service:
    metadata:
      annotations:
        my-annotation: value1
      labels:
        my-label: value2
    type: NodePort
    nodePort: 32000
```

For our `hippo` cluster, you would see the Service type and nodePort modification as well as the annotation and label.
For example:

```
kubectl -n ivory-operator get svc --selector=ivory-operator.ivorysql.org/cluster=hippo
```

will yield something similar to:

```
NAME              TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)          AGE
hippo-ha          NodePort    10.105.57.191   <none>        5432:32000/TCP   48s
hippo-ha-config   ClusterIP   None            <none>        <none>           48s
hippo-pods        ClusterIP   None            <none>        <none>           48s
hippo-primary     ClusterIP   None            <none>        5432/TCP         48s
hippo-replicas    ClusterIP   10.106.18.99    <none>        5432/TCP         48s
```

and the top of the output from running

```
kubectl -n ivory-operator describe svc hippo-ha
```

will show our custom annotation and label have been added:

```
Name:              hippo-ha
Namespace:         ivory-operator
Labels:            my-label=value2
                   ivory-operator.ivorysql.org/cluster=hippo
                   ivory-operator.ivorysql.org/patroni=hippo-ha
Annotations:       my-annotation: value1
```

Note that setting the `nodePort` value is not allowed when using the (default) `ClusterIP` type, and it must be in-range
and not otherwise in use or the operation will fail. Additionally, be aware that any annotations or labels provided here
will win in case of conflicts with any annotations or labels a user configures elsewhere.

Finally, if you are exposing your Services externally and are relying on TLS
verification, you will need to use the [custom TLS]({{< relref "tutorial/customize-cluster.md" >}}#customize-tls)
features of IVYO).

## Connect an Application

For this tutorial, we are going to connect [Keycloak](https://www.keycloak.org/), an open source
identity management application. Keycloak can be deployed on Kubernetes and is backed by a Ivory
database. While we provide an [example of deploying Keycloak and a ivorycluster](https://github.com/ivorysql/ivory-operator-examples/tree/main/kustomize/keycloak)
in the [Ivory Operator examples](https://github.com/ivorysql/ivory-operator-examples)
repository, the manifest below deploys it using our `hippo` cluster that is already running:

```
kubectl apply --filename=- <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: keycloak
  namespace: ivory-operator
  labels:
    app.kubernetes.io/name: keycloak
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: keycloak
  template:
    metadata:
      labels:
        app.kubernetes.io/name: keycloak
    spec:
      containers:
      - image: quay.io/keycloak/keycloak:latest
        args: ["start-dev"]
        name: keycloak
        env:
        - name: DB_VENDOR
          value: "ivory"
        - name: DB_ADDR
          valueFrom: { secretKeyRef: { name: hippo-pguser-hippo, key: host } }
        - name: DB_PORT
          valueFrom: { secretKeyRef: { name: hippo-pguser-hippo, key: port } }
        - name: DB_DATABASE
          valueFrom: { secretKeyRef: { name: hippo-pguser-hippo, key: dbname } }
        - name: DB_USER
          valueFrom: { secretKeyRef: { name: hippo-pguser-hippo, key: user } }
        - name: DB_PASSWORD
          valueFrom: { secretKeyRef: { name: hippo-pguser-hippo, key: password } }
        - name: KEYCLOAK_ADMIN
          value: "admin"
        - name: KEYCLOAK_ADMIN_PASSWORD
          value: "admin"
        - name: KC_PROXY
          value: "edge"
        ports:
        - name: http
          containerPort: 8080
        - name: https
          containerPort: 8443
        readinessProbe:
          httpGet:
            path: /realms/master
            port: 8080
      restartPolicy: Always
EOF
```

Notice this part of the manifest:

```
- name: DB_ADDR
  valueFrom: { secretKeyRef: { name: hippo-pguser-hippo, key: host } }
- name: DB_PORT
  valueFrom: { secretKeyRef: { name: hippo-pguser-hippo, key: port } }
- name: DB_DATABASE
  valueFrom: { secretKeyRef: { name: hippo-pguser-hippo, key: dbname } }
- name: DB_USER
  valueFrom: { secretKeyRef: { name: hippo-pguser-hippo, key: user } }
- name: DB_PASSWORD
  valueFrom: { secretKeyRef: { name: hippo-pguser-hippo, key: password } }
```

The above manifest shows how all of these values are derived from the `hippo-pguser-hippo` Secret. This means that we do not need to know any of the connection credentials or have to insecurely pass them around -- they are made directly available to the application!

Using this method, you can tie application directly into your GitOps pipeline that connect to Ivory without any prior knowledge of how IVYO will deploy Ivory: all of the information your application needs is propagated into the Secret!

## Next Steps

Now that we have seen how to connect an application to a cluster, let's learn how to create a [high availability Ivory]({{< relref "./high-availability.md" >}}) cluster!
