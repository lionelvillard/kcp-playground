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

- In the kind cluster, look for the physical namespace corresponding to `demo1` and do:

  ```shell
  kubectl logs -lapp=ping -n <ns corresponding to demo1>
  2022/11/14 22:27:04 ping succeeded
  2022/11/14 22:27:06 ping succeeded
  ```

  To find the physical namespace, run this command:
  ```shell
  kubectl get ns -oyaml | grep -A 5 demo1 
  kcp.dev/namespace-locator: '{"syncTarget":{"workspace":"root:kind","name":"kind","uid":"6a749f22-44af-43a3-8e4e-a8be9c31934a"},"workspace":"root:user1","namespace":"demo1"}'
  creationTimestamp: "2022-11-15T00:24:34Z"
  labels:
    internal.workload.kcp.dev/cluster: a02KrpZo4bFNfnSgocRl7IxHx2WnZsTaUzdOTg
    kubernetes.io/metadata.name: kcp-lk6uvb9e5ldp
  name: kcp-lk6uvb9e5ldp  <--- USE THIS 
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

- In the kind cluster, look for the namespace corresponding to `demo2-2` and do:

  ```shell
   kubectl logs -lapp=ping -n <ns corresponding to demo2-2>
   2022/11/14 22:37:18 ping succeeded
   2022/11/14 22:37:20 ping succeeded
  ```

### Workspace-to-Workspace isolation

- In `user2` workspace, create one namespace:

   ```shell
   kubectl kcp ws root:user2; kubectl create ns demo3
   ```
  
- Deploy the same `Ping` deployed in `user1`, targeting `pong.demo2-1.svc.cluster.local`, in workspace `user2`:

  ```shell
  ko apply -f config/demo2/ping.yaml -- -n demo3
  ```

- In the kind cluster, look for the physical namespace corresponding to `demo3` and do:

  ```shell
   kubectl logs -lapp=ping -n <ns corresponding to demo3>
   2022/11/15 00:19:45 ping failed: Get "http://pong.demo2-1.svc.cluster.local": dial tcp: lookup pong.demo2-1.svc.cluster.local on 10.96.144.49:53: no such host
   2022/11/15 00:19:47 ping failed: Get "http://pong.demo2-1.svc.cluster.local": dial tcp: lookup pong.demo2-1.svc.cluster.local on 10.96.144.49:53: no such host
  ```

### Pod-to-Pod isolation

- Look for `Pong` pod IP in workspace `user1`. Look for the physical namespace corresponding to `demo2-2`.

   ```shell
   kubectl get svc pong -ojsonpath='{.spec.clusterIP}'
   10.96.190.219
   ```
   
- Deploy the same `Ping`, this time targeting the service in workspace `user1` using the pod IP:

  ```shell
  kubectl kcp ws root:user2;\
  cat config/demo3/ping.yaml | sed "s/PODIP/http:\/\/10.96.190.219/" | ko apply -f - -- -n demo3
  ```

- In the kind cluster, look for the physical namespace corresponding to `demo3` and do:
  
  ```shell
  kubectl logs -lapp=ping -n <ns corresponding to demo3>
  2022/11/15 00:41:20 ping succeeded
  2022/11/15 00:41:22 ping succeeded
  ``` 
  
  Currently KCP does not isolate pods using network policies. See https://github.com/kcp-dev/kcp/issues/1988 for more details.
   
