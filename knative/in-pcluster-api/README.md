# Synchronizing Knative APIs in a physical cluster to kcp

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

## Create an API workspace exposing Knative APIs

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

- Then bind:

  ```shell
  kubectl kcp bind apiexport root:knative:serving
  ```

- Verify the Knative resources are available:

  ```shell
  kubectl api-resources | grep serving
  services                          kservice,ksvc   serving.knative.dev/v1            true         Service
  ````

At this point you should be able to deploy a Knative service:

```shell
kn service create hello \
--image gcr.io/knative-samples/helloworld-go \
--port 8080 \
--env TARGET=World
```

Since your home workspace is not bound to a location workspace the Knative service fails to become ready. Let's fix that.

## Create a Knative location workspace

The easiest way to install Knative is to use the [Knative quickstart installation method](https://knative.dev/docs/install/quickstart-install/).

- After installing the quickstart kn plugin, run this command (in a fresh new terminal):

  ```shell
  kn quickstart kind --install-serving
  ```

- Back to the kcp terminal, create a location workspace:

  ```shell
  kubectl kcp ws root
  kubectl kcp ws create kind-knative --enter
  ```

- Create a synctarget and ask to synchronize knative serving APIs (TODO: ideally only those APIs)

  ```shell
  kubectl kcp workload sync kind --resources=services.serving.knative.dev --syncer-image=ghcr.io/kcp-dev/kcp/syncer:main -o kind-syncer.yaml
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

## Bind your home workspace to the knative location workspace

- Go back to your home workspace:

  ```shell
  kubectl kcp ws
  ```

- And create a placement (TODO: currently it does not seem to be possible to just create a placement using the kcp command line)

  ```shell
  kubectl apply -f kind-placement.yaml
  ```

At this point I would expect the previous created Knative Service to be synchronized to the knative location but it's not happening.

As a workaround, delete `hello`:

  ```shell
  kn service delete hello
  ```

And recreate:

  ```shell
  kn service create hello \
  --image gcr.io/knative-samples/helloworld-go \
  --port 8080 \
  --env TARGET=World
  ```

## Delete the placement

I would expect the object to be deleted on the physical clusters.
I would expect new services to not be synchronized.



