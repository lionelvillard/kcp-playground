# Compute: bind k8s compute

The document describes the steps to 
1. create a workload workspace 
2. bind to a workload workspace 

## Setup

- In one terminal, start KCP:
  
  ```shell
  kcp start
  ```

- In another terminal, export KUBECONFIG to point to your KCP instance: 

   ```shell
   export KUBECONFIG=$(pwd)/.kcp/admin.kubeconfig
   ```
  
## Create a workload workspace

- Create a workspace of default type (from root: organization)

   ```shell
   kubectl kcp ws create kind --enter
   ```

  Output:

   ```{ .bash .no-copy }
   Workspace "kind" (type root:organization) created. Waiting for it to be ready...
   Workspace "kind" (type root:organization) is ready to use.
   Current workspace is "root:kind".
   ```

- Create a SyncTarget in the kind workspace. 

    ```shell
    kubectl kcp workload sync kind --syncer-image=ghcr.io/kcp-dev/kcp/syncer:main -o kind-syncer.yaml
    ```

  This command creates a SyncTarget, a Location and an APIExport.

  > For local dev, use kubectl kcp workload sync kind --syncer-image=kind.local/syncer  -o kind-syncer.yaml

- (optional) In another terminal, create a kind cluster.  
  Since the default kind CNI does not support network policies you need to disable it and install
  an CNI plugin support network policies. You can use Calico.

   ```shell
   kind create cluster --config kind/config-calico.yaml
   ```

   Then install Calico:

   ```shell
   kubectl apply -f https://docs.projectcalico.org/manifests/calico.yaml
   kubectl -n kube-system set env daemonset/calico-node FELIX_IGNORELOOSERPF=true
   ```
 
- Apply the syncer manifest to your kind cluster:

  ```shell
  kubectl apply -f kind-syncer.yaml
  ```

- Verify the SyncTarget is ready:

  ```shell
  kubectl get synctargets.workload.kcp.dev kind -ojsonpath='{.status.conditions[?(@.type=="Ready")].status}'
  True
  ```
- Verify the number of available Location is one:

  ```shell
  kubectl get locations.scheduling.kcp.dev default
  NAME      RESOURCE      AVAILABLE   INSTANCES   LABELS   AGE
  default   synctargets   1           1                    6m55s
  ```

  The `kind` workspace is a compute workspace. By default, Deployments, Services
and Ingresses are imported from the location:

   ```shell 
    kubectl get apiresourceimports
    NAME
    deployments.kind.v1.apps
    ingresses.kind.v1.networking.k8s.io
    services.kind.v1.core
   ```

Let's import these resources in your workspace.

## Bind to a user workspace

- Create a workspace of default type (from organiation: universal)

   ```shell
   kubectl kcp ws create mine --enter
   ```

  Output:

   ```{ .bash .no-copy }
   Workspace "mine" (type root:universal) created. Waiting for it to be ready...
   Workspace "mine" (type root:universal) is ready to use.
   Current workspace is "root:kind:mine".
   ```
  
Note: the workspace does not have to be a child of `kind`

- Bind compute (ie. Deployments, Services and Ingresses):

   ```shell
   kubectl kcp bind compute root:kind
   ```

   Output:

  ```shell
  apibinding kubernetes-1pre20xf for apiexport root:compute:kubernetes created.
  placement placement-2iddvmcj created. 
  ``` 

- Create a deployment in the user1 workspace

  ```shell 
  kubectl create deployment --image=gcr.io/kuar-demo/kuard-amd64:blue --port=8080 kuard
  ```

  Output:

  ```shell
  deployment.apps/kuard created
  ```
