# Getting Started
1. Installed IVYO to the `ivory-operator` namespace. If you are inside your `examples` directory, you can run the `kubectl apply --server-side -k examples/kustomize/install/default` command.

Throughout this tutorial, we will be building on the example provided in the `examples/kustomize/ivory`.

When referring to a nested object within a YAML manifest, we will be using the `.` format similar to `kubectl explain`. For example, if we want to refer to the deepest element in this yaml file:

```
spec:
  hippos:
    appetite: huge
```

we would say `spec.hippos.appetite`.

`kubectl explain` is your friend. You can use `kubectl explain ivorycluster` to introspect the `ivorycluster.ivory-operator.ivorysql.org` custom resource definition.

With IVYO, the Ivory Operator installed, let's go and [create a Ivory cluster](https://github.com/IvorySQL/ivory-operator/blob/master/docs/content/tutorial/create-cluster.md)!
