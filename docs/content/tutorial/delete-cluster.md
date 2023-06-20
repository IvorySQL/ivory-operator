---
title: "Delete a Ivory Cluster"
date:
draft: false
weight: 110
---

There comes a time when it is necessary to delete your cluster. If you have been [following along with the example](https://github.com/ivorysql/ivory-operator-examples), you can delete your Ivory cluster by simply running:

```
kubectl delete -k kustomize/ivory
```

IVYO will remove all of the objects associated with your cluster.

With data retention, this is subject to the [retention policy of your PVC](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#reclaiming). For more information on how Kubernetes manages data retention, please refer to the [Kubernetes docs on volume reclaiming](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#reclaiming).
