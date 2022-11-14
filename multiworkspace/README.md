# Virtual Workspaces

The document shows:
1. services deployed in a workspace can communicate with each others.
2. services deployed in another workspace bound to the same 
location (or workload) workspace cannot communicate with the services created in 
the first workspace.

## Setup

### KCP Setup

see [KCP setup](../compute/README.md#setup).

### Create a workload workspace

See [Create a workload workspace](../compute/README.md#Create-a-workload-workspace).

### Create two workspaces

- Create `user1` workspace

   ```shell
   kubectl kcp ws root; kubectl kcp ws create user1 --enter
   ```

- Bind to the workload workspace:

   ```shell
   kubectl kcp bind compute root:kind
   ```
 
- Observe API resource availability in user1 workspace:

  ```{ shell .no-copy }
  kubectl api-resources | grep deploy
  deployments   deploy   apps/v1    true         Deployment
  ```

- Create `user2` workspace

   ```shell
   kubectl kcp ws root; kubectl kcp ws create user2 --enter
   ```

- Bind to the workload workspace:

   ```shell
   kubectl kcp bind compute root:kind
   ```

- Observe API resource availability in user1 workspace:

  ```{ shell .no-copy }
  kubectl api-resources | grep deploy
  deployments   deploy   apps/v1    true         Deployment
  ```

## Scenarios

### Svc-to-Svc communication, same namespace

- In `user1` workspace, create a namespace
   
   ```shell
   kubectl kcp ws root:user1; kubectl create ns demo1 
   ```

- Deploy `Ping` and `Pong` services:

  ```shell
  ko apply -f config/demo1/ -- -n demo1
  ```

- In the kind cluster, look for the namespace corresponding to `demo1` and do:

  ```shell
  kubectl logs -lapp=ping -n <ns corresponding to demo1>
  2022/11/14 22:27:04 ping succeeded
  2022/11/14 22:27:06 ping succeeded
  ```

### Svc-to-Svc communication, different namespaces

- In `user1` workspace, create one namespace:

   ```shell
   kubectl kcp ws root:user1; kubectl create ns demo2-1
   ```

- And another one:

   ```shell
   kubectl kcp ws root:user1; kubectl create ns demo2-2
   ```

- Deploy `Pong` in namespace demo2-1:

  ```shell
  ko apply -f config/demo2/pong.yaml -- -n demo2-1
  ```

- Deploy `Ping` in namespace demo2-2:

  ```shell
  ko apply -f config/demo2/ping.yaml -- -n demo2-2
  ```
- 
- In the kind cluster, look for the namespace corresponding to `demo2-2` and do:

  ```shell
   kubectl logs -lapp=ping -n <ns corresponding to demo2-2>
   2022/11/14 22:37:18 ping succeeded
   2022/11/14 22:37:20 ping succeeded
  ```


### Workspace-to-Workspace isolation

TBA.
