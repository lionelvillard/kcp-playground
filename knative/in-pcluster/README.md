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

## Create a workspace and synchronize Knative APIs


- Create a workspace:

  ```shell
  kubectl kcp ws create knative-app --enter
  ```

- Create the syncer manifest:

  ```shell
  kubectl kcp workload sync kind --resources=services.serving.knative.dev --syncer-image=ghcr.io/kcp-dev/kcp/syncer:main -o kind-syncer.yaml
  ```

- Apply the syncer manifest to your kind cluster:

  ```shell
  kubectl apply -f kind-syncer.yaml
  ```

- Verify the SyncTarget is ready:

  ```shell
  kubectl get synctargets.workload.kcp.io kind -ojsonpath='{.status.conditions[?(@.type=="Ready")].status}'
  True
  ```

- Verify the Knative API has been imported into your worskspace:

  ```shell
  kubectl api-resources | grep knative
  services                          kservice,ksvc   serving.knative.dev/v1            true         Service
  ```

- Place root:knative-app in root:knative-app location

  ```shell
  kubectl kcp bind compute root:knative-app
  ```

## Testing

Deploy a Knative service in your kcp workspace:

```shell
kn service create hello \
--image gcr.io/knative-samples/helloworld-go \
--port 8080 \
--env TARGET=World
```

After a while, observe the service is ready:

```shell
kubectl get ksvc
NAME    URL                                                LATESTCREATED   LATESTREADY   READY   REASON
hello   http://hello.kcp-5tyerk3vym4m.127.0.0.1.sslip.io   hello-00001     hello-00001   True    
```

Then curl it:

```shell
curl http://hello.kcp-5tyerk3vym4m.127.0.0.1.sslip.io
Hello World!
```


