# Compute: bind k8s compute

This is a getting started guide about kcp. It shows how to create location workspaces and how to consume (bind) those workspaces inside your user workspace.

## Running kcp locally

Assuming you have already installed kcp on your computer, run
the following commands.

- In one terminal, start kcp:

  ```shell
  kcp start
  ```

- In another terminal, export KUBECONFIG to point to your KCP instance:

   ```shell
   export KUBECONFIG=$(pwd)/.kcp/admin.kubeconfig
   ```

## Creating a location workspace

- Create a *location workspace* of default type under root (root:organization)

   ```shell
   kubectl ws root
   kubectl ws create local --enter
   ```

  Output:

   ```{ .bash .no-copy }
   Workspace "local" (type root:organization) created. Waiting for it to be ready...
   Workspace "local" (type root:organization) is ready to use.
   Current workspace is "root:local".
   ```

- Create a SyncTarget in the local workspace.

    ```shell
    kubectl kcp workload sync local --syncer-image=ghcr.io/kcp-dev/kcp/syncer:main -o local-syncer.yaml
    ```

  > For local dev using Kind, use `kubectl kcp workload sync local --syncer-image=kind.local/syncer  -o local-syncer.yaml`, assuming using `ko` to build the syncer image and `KO_DOCKER_REPO` is set to `kind.local`

  This command creates:
  - a SyncTarget object representing the physical cluster
  - a default Location object with local SyncTarget as being the sole target
  - an APIBinding object importing root:kubernetes APIs
  - an APIExport object reexporting kubernetes APIs (and any additional resources specified in the `--resources` command flag. See below for more details).
  - a manifest file installing the syncer to the physical cluster

- Apply the syncer manifest to your (local) cluster (make sure KUBECONFIG points to your physical cluster):

  ```shell
  kubectl apply -f local-syncer.yaml
  ```

- Verify the SyncTarget (in kcp) is ready, indicating the syncer deployment in the physical cluster is healthy and has successfully communicating its status back to kcp:

  ```shell
  kubectl get synctargets.workload.kcp.io local -ojsonpath='{.status.conditions[?(@.type=="Ready")].status}'
  ```

  ```shell
  True
  ```

- Verify the number of available Location is one:

  ```shell
  kubectl get locations.scheduling.kcp.io default
  NAME      RESOURCE      AVAILABLE   INSTANCES   LABELS   AGE
  default   synctargets   1           1                    6m55s
  ```

The SyncTarget imports in kcp resources available on the physical clusters, by default deployments, ingresses, services and pods:

   ```shell
    kubectl get apiresourceimports
    NAME
    deployments.local.v1.apps
    ingresses.local.v1.networking.k8s.io
    pods.local.v1.core
    services.local.v1.core
   ```

  > To import additional resources, add `--resource=<comma-separated resource names` to the `kcp worload sync` command

Let's import these resources in your user workspace.

## Consuming location workspaces

- Go to your home workspace:

   ```shell
   kubectl kcp ws
   ```

   ```{ .bash .no-copy }
   Current workspace is "kvdk2spgmbix".
   ```

- Bind the location workspace previous created using `kcp bind compute`:

   ```shell
   kubectl kcp bind compute root:local
   ```

   Output:

  ```shell
  Binding APIExport "root:compute:kubernetes".
  placement placement-ej377u0k created.
  Placement "placement-ej377u0k" is ready.
  ```

  This command creates two objects:
  - an APIBinding object importing the standard kubernetes APIs exported by the `root:compute` workspace
  - a Placement object indicating that all objects, in all namespaces in your home workspace are placed in the location workspace `root:local`.

- Create a deployment:

  ```shell
  kubectl create deployment --image=gcr.io/kuar-demo/kuard-amd64:blue --port=8080 kuard
  ```

  Output:

  ```shell
  deployment.apps/kuard created
  ```

Voila!
