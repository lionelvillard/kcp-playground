# Providing physical Knative as a Service

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

## Create a kind cluster with Knative

The easiest way to install Knative is to use the [Knative quickstart installation method](https://knative.dev/docs/install/quickstart-install/).

After installing the quickstart kn plugin, run this command (in a fresh new terminal):

```shell
kn quickstart kind --install-serving
```

This command creates a kind cluster named `knative` and installs Knative Serving.

## Create a workspace holding Knative artifacts

- Create a workspace:

  ```shell
  kubectl kcp ws create knative --enter
  ```

- Apply the Knative resource schema:

  ```shell
  kubectl apply -f service-ars.yaml
  ```

- And export:

  ```shell
  kubectl apply -f serving-apiexport.yaml
  ```

## Bind the Knative Serving API in your home workspace

- Go to your home workspace:

  ```shell
  kubectl kcp ws
  ```

- Then bind to the API:

  ```shell
  kubectl kcp bind apiexport root:knative:serving
  ```

- Verify the Knative resources are available:

  ```shell
  kubectl api-resources | grep serving
  services                          kservice,ksvc   serving.knative.dev/v1            true         Service
  ````

At this point you should be able to create a Knative service:

```shell
kn service create hello \
--image gcr.io/knative-samples/helloworld-go \
--port 8080 \
--env TARGET=World
```

Since your home workspace is not bound to a location workspace the Knative service fails to become ready. Let's fix that.

## Add a Knative location

The easiest way to install Knative to a physical cluster is to use the [Knative quickstart installation method](https://knative.dev/docs/install/quickstart-install/).

- After installing the quickstart kn plugin, run this command (in a fresh new terminal):

  ```shell
  kn quickstart kind --install-serving
  ```

- Back to the kcp terminal, switch the knative workspace:

  ```shell
  kubectl kcp ws root:knative
  ```

- Create a SyncTarget and ask to synchronize knative serving APIs only:

  ```shell
  kubectl kcp workload sync kind --apiexports=root:knative:serving --syncer-image=ghcr.io/kcp-dev/kcp/syncer:main -o kind-syncer.yaml
  ```

- Apply the syncer manifest into your kind cluster:

  ```shell
  kubectl apply -f kind-syncer.yaml
  ```

- Verify the SyncTarget is ready:

  ```shell
  kubectl get synctargets.workload.kcp.io kind -ojsonpath='{.status.conditions[?(@.type=="Ready")].status}'
  True
  ```

- Verify all bound resources are compatible with the resources available in the physical cluster:

  ```shell
  kubectl get synctargets.workload.kcp.io kind -oyaml
  ```

  ```yaml
  ...
   syncedResources:
   - group: serving.knative.dev
     identityHash: 714b9f57a70690e65ccbce50fea9f01efad00573d5da03c8f3a9feb3ff5d9ca6
     resource: services
     state: Accepted
     versions:
     - v1
  ...
  ```

## Bind your home workspace to the knative location workspace

- Go back to your home workspace:

  ```shell
  kubectl kcp ws
  ```

- And create a placement:

  ```shell
  kubectl kcp bind compute root:knative --apiexports=root:knative:serving --name location1
  ```


- Wait a bit and you should see the `hello` service is ready:

  ```shell
  kubectl get ksvc hello
  ```

  ```shell
  NAME    URL                                                LATESTCREATED   LATESTREADY   READY   REASON
  hello   http://hello.kcp-2hs1yezgzr6n.127.0.0.1.sslip.io   hello-00001     hello-00001   True
  ```

## Delete the placement

- Delete the placement object:

  ```shell
  kubectl delete placement location1
  ```

  Wait a bit (30s by default) and observe in the physical cluster the objects have been deleted.
