# Multi-Workspaces

- Start KCP
  
  ```shell
  kcp start
  ```

- export KUBECONFIG

   ```shell
   export KUBECONFIG=$(pwd)/.kcp/admin.kubeconfig
   ```

- Create a workspace. 

   ```shell
   kubectl kcp ws create knative --enter
   ```

   Output:

   ```{ .bash .no-copy }
   Workspace "knative" (type root:organization) created. Waiting for it to be ready...
   Workspace "knative" (type root:organization) is ready to use.
   Current workspace is "root:knative".
   ```

- Prepare for the Syncer installation. This command needs to run in the `knative` workspace:

  - Syncer Dev:
    ```shell
    kubectl kcp workload sync kind --syncer-image=kind.local/syncer:latest -o kind-syncer.yaml
    ```
  - Prod:
    ```shell
    kubectl kcp workload sync kind --syncer-image=ghcr.io/kcp-dev/kcp/syncer:fe25bb1 -o kind-syncer.yaml
    ```

- Apply the syncer manifest to your kind cluster:
  
  ```shell
  kubectl apply -f kind-syncer.yaml
  ```

- Bind compute (ie. Deployments, Services and Ingresses):

   ```shell
   kubectl kcp bind compute root:knative 
   ```

- Convert Knative Service CRD to APIResourceSchema and apply it :

    ```shell
    kubectl kcp crd snapshot -f services-crd.yaml --prefix v1 | kubectl apply -f -
    ```

- Create an APIExport in the knative organization:

  ```shell
  cat <<EOF | kubectl apply -f - 
  apiVersion: apis.kcp.dev/v1alpha1 
  kind: APIExport 
  metadata: 
    name: knative 
  spec: 
    latestResourceSchemas:
    - v1.services.serving.knative.dev 
  EOF
  ```

- Create a `user1` workspace

   ```shell
   kubectl kcp ws ..
   kubectl kcp ws create user1 --enter
   ```

- Create a binding in user1 workspace to the APIExport in the organisation workspace

   ```shell
   cat << EOF | kubectl apply -f - 
   apiVersion: apis.kcp.dev/v1alpha1
   kind: APIBinding
   metadata:
     name: knative
   spec:
     reference:
       workspace:
         path: "root:knative"
         exportName: knative
   EOF
   ```

- Observe API resource availability in user1 workspace:

  ```{ shell .no-copy }
  kubectl api-resources | grep knative
  services                          kservice,ksvc   serving.knative.dev/v1            true         Service
  ```

- Create a `user2` workspace and a binding inside:

   ```shell
   kubectl ws ..
   kubectl ws create user2 --enter
   ```

   ```shell
   cat << EOF | kubectl apply -f -  
   apiVersion: apis.kcp.dev/v1alpha1
   kind: APIBinding
   metadata:
     name: kubernetes
   spec:
     reference:
       workspace:
         path: "root:knative"
         exportName: knative
   EOF
   ```


- Create a deployment in the user1 workspace

```shell
kubectl kcp ws root:user1
kubectl create deployment --image=gcr.io/kuar-demo/kuard-amd64:blue --port=8080 kuard
```

