# Virtual Workspaces

The document describes the steps to create virtual workspaces.

## Setup
 
See [setup](../compute/README.md#setup).


## Create a workload workspace

See [Create a workload workspace](../compute/README.md#Create-a-workload-workspace).


## Create a virtual workspace

- Create a `user1` workspace

   ```shell
   kubectl kcp ws ..
   kubectl kcp ws create user1 --enter
   ```

- Bind to the kind location:

   ```shell
   kubectl kcp bind compute root:kind
   ```
 
- Observe API resource availability in user1 workspace:

  ```{ shell .no-copy }
  kubectl api-resources | grep deploy
  deployments   deploy   apps/v1    true         Deployment
  ```

- Create a deployment in the user1 workspace
  
  ```shell
  kubectl create deployment --image=gcr.io/kuar-demo/kuard-amd64:blue --port=8080 kuard
  ```

- Create a `user2` workspace and a binding:

   ```shell
   kubectl ws ..
   kubectl ws create user2 --enter
   kubectl kcp bind compute root:kind
   ```
- 
- Create a deployment in the user2 workspace

  ```shell
  kubectl create deployment --image=gcr.io/kuar-demo/kuard-amd64:blue --port=8080 kuard
  ```


