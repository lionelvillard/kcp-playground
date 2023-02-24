# Views Isolation (aka Virtual Workspaces)

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

Make sure the workload workspace is called `root:kind`

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

### Breaking Workspace-to-Workspace isolation

This scenario shows how a pod in workspace `user1` can attempt to resolve
the IP of a pod in workspace `user2` to send requests to it.

- In `user1` workspace, create one namespace:

   ```shell
   kubectl kcp ws root:user1; kubectl create ns demo-bad-w2w-1
   ```

- In `user2` workspace, create one namespace:

   ```shell
   kubectl kcp ws root:user2; kubectl create ns demo-bad-w2w-2
   ```

- Deploy `Pong` in `user2`:

  ```shell
  ko apply -f config/demo4/pong.yaml -- -n demo-bad-w2w-2
  ```

- Get the DNS IP of the DNS pod resolving `user2` addresses. Finding which DNS pod corresponds
  to `user2` is a bit tricky but feasible by looking at the DNS `ConfigMap`. In the syncer namespace in the pcluster, do:

  ```shell
  kubectl get cm -oyaml
  ```

  then look for `demo-bad-w2w-2` and get the name of the ConfigMap containing it. This is also the name
  of the DNS Service. Do:

  ```shell
  kubectl get svc kcp-dns-kind-7rohw1hj-2c88xuxg # replace by name found above
  NAME                             TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)         AGE
  kcp-dns-kind-qwaqgajv-2plia53u   ClusterIP   10.96.10.16   <none>        53/UDP,53/TCP   15h
  ```

  CLUSTER-IP is the IP you will use below.

- Deploy `Ping-attack` in `user1`:

  ```shell
  kubectl kcp ws root:user1;\
  cat config/demo4/ping.yaml | sed "s/REPLACE/10.96.10.16:53/" | ko apply -f - -- -n demo-bad-w2w-1
  ```

- In the ping log (with network policy PR):

  ```shell
  dialing to 10.96.10.16:53
  2022/11/29 17:32:30 ping failed: Get "http://pong.demo-bad-w2w-2.svc.cluster.local": dial tcp: lookup pong.demo-bad-w2w-2.svc.cluster.local on 10.96.174.212:53: read udp 192.169.82.9:40943->10.96.10.16:53: read: connection refused
  ```

- Delete the network policies:

  ```shell
  kubectl delete -n kcp-syncer-kind-1wd274wf networkpolicies.networking.k8s.io --all
  ```

- In the ping log (it can take a while for the networkpolicies to apply):

  ```shell
  dialing to 10.96.10.16:53
  2022/11/29 17:34:15 ping succeeded
  ```

### Pod-to-Pod isolation

- Look for `Pong` pod IP in workspace `user1`. Look for the physical namespace corresponding to `demo2-1`.

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

