# Quickstart

> You need to have a Kubernetes cluster, and the `kubectl` must be configured to communicate with your cluster.

The simplest setup of module-sdk in your cluster consists of these steps:

- build an image with your module
- create necessary RBAC objects (for Kubernetes access)
- deploy your module to the cluster

For more configuration options see [RUNNING](RUNNING.md).

## Build an image with your module

A module is a component that implements specific functionality using the module-sdk framework. [Learn more](MODULE_DEVELOPMENT.md) about module development.

Let's create a small module that will watch for all Pods in all Namespaces and simply log the name of a new Pod.

Create a basic module structure with a handler that responds to Pod creation events. Create the `pod-watcher.go` file with the following content:

```go
package main

import (
  "fmt"
  "github.com/your-org/module-sdk/pkg/module"
  corev1 "k8s.io/api/core/v1"
  "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
  "k8s.io/client-go/kubernetes/scheme"
)

func main() {
  mod := module.New("pod-watcher")
  
  mod.RegisterEventHandler("v1", "Pod", module.EventHandlerConfig{
    EventTypes: []string{"Added"},
    Handler: func(obj *unstructured.Unstructured) error {
      // Convert unstructured to Pod
      pod := &corev1.Pod{}
      err := scheme.Scheme.Convert(obj, pod, nil)
      if err != nil {
        return err
      }
      
      fmt.Printf("Pod '%s' added\n", pod.Name)
      return nil
    },
  })
  
  mod.Run()
}
```

Create the following `Dockerfile` in the directory where you created the `pod-watcher.go` file:

```dockerfile
FROM golang:1.19 as builder
WORKDIR /app
COPY . .
RUN go mod init pod-watcher && \
  go mod tidy && \
  CGO_ENABLED=0 go build -o /pod-watcher

FROM alpine:3.16
COPY --from=builder /pod-watcher /pod-watcher
ENTRYPOINT ["/pod-watcher"]
```

Build an image (change image tag according to your Docker registry):

```sh
docker build -t "registry.mycompany.com/module-sdk:pod-watcher" .
```

Push image to the Docker registry accessible by the Kubernetes cluster:

```sh
docker push registry.mycompany.com/module-sdk:pod-watcher
```

## Create RBAC objects

We need to watch for Pods in all Namespaces. That means that we need specific RBAC definitions for our module:

```sh
kubectl create namespace example-pod-watcher
kubectl create serviceaccount pod-watcher-acc --namespace example-pod-watcher
kubectl create clusterrole pod-watcher --verb=get,watch,list --resource=pods
kubectl create clusterrolebinding pod-watcher --clusterrole=pod-watcher --serviceaccount=example-pod-watcher:pod-watcher-acc
```

## Deploy your module to the cluster

Module-sdk modules can be deployed as a Pod or Deployment. Put this manifest into the `pod-watcher.yaml` file:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: pod-watcher
spec:
  containers:
  - name: pod-watcher
  image: registry.mycompany.com/module-sdk:pod-watcher
  imagePullPolicy: Always
  serviceAccountName: pod-watcher-acc
```

Deploy your module by applying the `pod-watcher.yaml` file:

```sh
kubectl -n example-pod-watcher apply -f pod-watcher.yaml
```

## It all comes together

Let's deploy a [kubernetes-dashboard][kubernetes-dashboard] to trigger the registered event handler in our module:

```sh
kubectl apply -f https://raw.githubusercontent.com/kubernetes/dashboard/master/aio/deploy/recommended.yaml
```

Now run `kubectl -n example-pod-watcher logs po/pod-watcher` and observe that the module will print dashboard pod name:

```plain
...
2023/06/22 12:30:45 Registered event handler for v1/Pod
2023/06/22 12:30:45 Starting informers...
2023/06/22 12:30:50 Pod 'kubernetes-dashboard-775dd7f59c-hr7kj' added
...
```

To clean up a cluster, delete namespace and RBAC objects:

```sh
kubectl delete ns example-pod-watcher
kubectl delete clusterrole pod-watcher
kubectl delete clusterrolebinding pod-watcher
```

This example is also available in /examples: [pod-watcher-example][pod-watcher-example].

[kubernetes-dashboard]: https://kubernetes.io/docs/tasks/access-application-cluster/web-ui-dashboard/
[pod-watcher-example]: https://github.com/your-org/module-sdk/tree/main/examples/pod-watcher
