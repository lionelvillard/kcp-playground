# Installing Knative Serving in KCP

This guide describes the steps to install Knative Serving in [kcp](http://kcp.io).

Status: some kcp features are still missing for this guide to be completed. WIP.

## Prerequisites

- A Kubernetes cluster.
- ko (optional)
- kcp source code clone locally and https://github.com/kcp-dev/kcp/pull/2910 checked out.

## Clone this repository

In a directory:

```shell
git clone https://github.com/lionelvillard/kcp-playground.git
cd kcp-playground/knative/in-kcp
```

## Preparing kcp

1. In a terminal, start kcp:

   ```shell
   kcp start
   ```

1. In the kcp terminal, make sure your KUBECONFIG points to your kcp instance:

   ```shell
   export KUBECONFIG=$(pwd)/.kcp/admin.kubeconfig
   ```

1. Create an organization workspace called `knative` and immediately enter it:

    ```shell
    kubectl kcp workspace create knative --enter
    ```

    ```shell
    Workspace "knative" (type root:organization) created. Waiting for it to be ready...
    Workspace "knative" (type root:organization) is ready to use.
    Current workspace is "root:knative".
    ```

## Registering Kind as a SyncTarget

1. In a new terminal, create a Kind cluster configured to map Knative ingress to localhost port 80:

   ```shell
   kind create cluster --config kind/config-knative.yaml
   ```

1. (optional) Build the syncer image from kcp source. if not replace `kind.local/syncer` below to point to a released syncer image (eg. `ghcr.io/kcp-dev/kcp/syncer:main`).

   ```shell
   cd <wherever is kcp>
   export KO_DOCKER_REPO=kind.local
   ko build -B ./cmd/syncer
   ```

1. In the kcp terminal, create a SyncTarget:

    ```shell
    kubectl kcp workload sync kind --resources=poddisruptionbudgets.policy,horizontalpodautoscalers.autoscaling --syncer-image kind.local/syncer -o kind-syncer.yaml
    ```

1. Register the kind cluster:

    ```shell
    kubectl apply -f kind-syncer.yaml
    ```

1. Back to the terminal pointing to kcp, verify the syncer is ready. Run this command against kcp:

    ```shell
    kubectl get synctargets.workload.kcp.io kind -ojsonpath='{.status.conditions[?(@.type=="Ready")].status}'
    True
    ```

1. Finally, bind compute (ie. Deployments, Services, Ingresses and Pods):

   ```shell
   kubectl kcp bind compute root:knative
   ```

## Installing Knative

1. Make sure kubectl points to the kcp knative workspace

   ```shell
   kubectl kcp ws .
   Current workspace is "root:knative".
   ```

1. Install the Knative CRDs in the kcp knative workspace

    ```shell
    kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.9.2/serving-crds.yaml
    ```

1. Verify

    ```shell
    kubectl get crds
    ```

    ```
    NAME                                                  CREATED AT
    certificates.networking.internal.knative.dev          2022-07-05T21:26:59Z
    clusterdomainclaims.networking.internal.knative.dev   2022-07-05T21:26:59Z
    configurations.serving.knative.dev                    2022-07-05T21:26:59Z
    domainmappings.serving.knative.dev                    2022-07-05T21:26:59Z
    images.caching.internal.knative.dev                   2022-07-05T21:27:00Z
    ingresses.networking.internal.knative.dev             2022-07-05T21:26:59Z
    metrics.autoscaling.internal.knative.dev              2022-07-05T21:26:59Z
    podautoscalers.autoscaling.internal.knative.dev       2022-07-05T21:27:00Z
    revisions.serving.knative.dev                         2022-07-05T21:27:00Z
    routes.serving.knative.dev                            2022-07-05T21:27:00Z
    serverlessservices.networking.internal.knative.dev    2022-07-05T21:27:00Z
    services.serving.knative.dev                          2022-07-05T21:27:00Z
    ```

   You should see only `knative.dev` CRDs.

1. Install Knative Serving Core:

    ```shell
    kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.9.2/serving-core.yaml
    ```

1. Annotate `activator-service` indicating kcp to synchronize derived endpoints to kcp:

   ```shell
   kubectl annotate -n knative-serving svc activator-service experimental.workload.kcp.io/upsync-derived-resources=endpoints
   ```

1. Wait a bit (20s-40s or more) and verify all Knative Serving deployments are ready:

   ```shell
   kubectl -n knative-serving get deployments.apps
   ```

   ```shell
   kubectl -n knative-serving get deployments.apps
   NAME                    READY   UP-TO-DATE   AVAILABLE   AGE
   activator               1/0     1            1           83s
   autoscaler              1/1     1            1           83s
   controller              1/0     1            1           83s
   domain-mapping          1/0     1            1           83s
   domainmapping-webhook   1/0     1            1           83s
   webhook                 1/0     1            1           82s
   ```

   > Note: there is a kcp bug setting default `replicas` to 0 instead of 1

1. Install the networking layer. This guide uses [net-kourier](https://github.com/knative-sandbox/net-kourier).

   ```shell
   kubectl apply -f https://github.com/knative/net-kourier/releases/download/knative-v1.9.2/kourier.yaml
   ```

1. Annotate `kourier-internal` indicating kcp to synchronize derived endpoints to kcp:

   ```shell
   kubectl annotate -n kourier-system svc kourier-internal experimental.workload.kcp.io/upsync-derived-resources=endpoints
   ```

1. Patch the network configmap:

   ```shell
   kubectl patch configmap/config-network \
           --namespace knative-serving \
           --type merge \
           --patch '{"data":{"ingress-class":"kourier.ingress.networking.knative.dev"}}'
   ```

1. Verify kourier is up and running:

   ```shell
   kubectl get deployments.apps -n kourier-system
   NAME                     READY   UP-TO-DATE   AVAILABLE   AGE
   3scale-kourier-gateway   1/0     1            1           98s
   ```

   ```shell
   kubectl get deployments.apps -n knative-serving net-kourier-controller
   NAME                     READY   UP-TO-DATE   AVAILABLE   AGE
   net-kourier-controller   1/1     1            1           2m45s
   ```

1. Setup magic DNS:

```shell
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Service
metadata:
  name: kourier-ingress
  namespace: kourier-system
  labels:
   networking.knative.dev/ingress-provider: kourier
spec:
  type: NodePort
  selector:
    app: 3scale-kourier-gateway
  ports:
  - name: http2
    nodePort: 31080
    port: 80
    targetPort: 8080
EOF
```

```shell
kubectl  patch configmap -n knative-serving config-domain \
  -p '{"data": {"127.0.0.1.sslip.io": ""}}'
```

## Deploying your first Knative service

In the KCP terminal, deploy the hello world app using kn:

   ```shell
   kn service create hello \
   --image gcr.io/knative-samples/helloworld-go \
   --port 8080 \
   --env TARGET=World \
   --annotation-revision experimental.workload.kcp.io/upsync-derived-resources=endpoints
   ``````

```
kn service create hello \
   --image gcr.io/knative-samples/helloworld-go \
   --port 8080 \
   --env TARGET=World \
   --annotation-revision experimental.workload.kcp.io/upsync-derived-resources=endpoints
Warning: Kubernetes default value is insecure, Knative may default this to secure in a future release: spec.template.spec.containers[0].securityContext.allowPrivilegeEscalation, spec.template.spec.containers[0].securityContext.capabilities, spec.template.spec.containers[0].securityContext.runAsNonRoot, spec.template.spec.containers[0].securityContext.seccompProfile
Creating service 'hello' in namespace 'default':

  1.083s The Configuration is still working to reflect the latest desired specification.
  2.070s Configuration "hello" is waiting for a Revision to become ready.
 39.716s ...
 40.141s Ingress has not yet been reconciled.
 40.524s Waiting for load balancer to be ready
 40.801s Ready to serve.

Service 'hello' created to latest revision 'hello-00001' is available at URL:
http://hello.default.127.0.0.1.sslip.io
```

Note that:
  -  `experimental.workload.kcp.io/upsync-derived-resources: endpoints` must be set so that Knative can detect the service is ready


## Testing

Since the `hello` Knative service has been installed locally using Kind, you can test the service with curl:

   ```shell
   curl http://hello.default.127.0.0.1.sslip.io
   Hello World!
   ```

## TODOs

TODOs:
- Eventing
- Deleting service
