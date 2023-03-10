# KCP Playground

This repository contains a collections of guides, tutorial and demos related to [kcp](http://kcp.io).

- [How to create a deployment in kcp that is synced to a physical cluster](./compute/README.md).
- [Demo showing multi-workspaces isolation guarantee](./multiworkspace/README.md)
- [Guide on how to export CRDs](./exportcrd/README.md)
- Various ways to expose Knative APIs:
  - [From existing physical Knative cluster](./knative/in-pcluster/README.md)
  - [Late binding: export Knative APIs and then bind to a Knative cluster](./knative/in-pcluster-api/README.md)
  - [By installing Knative in kcp](./knative/in-kcp/README.md)

## Kubernetes Cluster Setup

There are many different ways to create a Kubernetes cluster locally. Here are some examples.

### Kind (all platforms)

- Make sure [kind](https://kind.sigs.k8s.io) is installed on your machine.

- Create a k8s 1.24 cluster with network policies enabled:

   ```shell
   kind create cluster --config kind/config-calico.yaml
   ```
- Then install Calico:

   ```shell
   kubectl apply -f https://raw.githubusercontent.com/projectcalico/calico/release-v3.24/manifests/calico.yaml \
   kubectl -n kube-system set env daemonset/calico-node FELIX_IGNORELOOSERPF=true
   ```

### k3d (all platforms)

- Make sure [k3d](https://k3d.io/v5.4.8/#installation) is installed on your machine.

- Create a cluster named `workload`:

    ```shell
    k3d cluster create workload --image rancher/k3s:v1.24.10-k3s1
    ```


