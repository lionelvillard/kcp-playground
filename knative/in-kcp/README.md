# Installing Knative Serving in KCP

This guide describes the steps to install Knative Serving in [kcp](http://kcp.io).

Status: some kcp features are still missing for this guide to be completed. WIP.

## Prerequisites

- A Kubernetes cluster.
- ko (optional)
- kcp source code clone locally and https://github.com/kcp-dev/kcp/pull/2829 checked out.

## Clone this repository

In a directory:

```shell
git clone https://github.com/lionelvillard/kcp-playground.git
cd kcp-playground/knative/in-kcp
```

## Preparing kcp

1. In a terminal, start kcp with pod synchronization enabled:

   ```shell
   kcp start --feature-gates KCPSyncerTunnel=true
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

1. In a new terminal, create a Kind cluster:

   ```shell
   kind create cluster
   ```

1. (optional) Build the syncer image from kcp source. if not replace `kind.local/syncer` below to point to a released syncer image (eg. `ghcr.io/kcp-dev/kcp/syncer:main`).

   ```shell
   cd <wherever is kcp>
   export KO_DOCKER_REPO=kind.local
   ko build -B ./cmd/syncer
   ```

1. In the kcp terminal, create a SyncTarget with syncer tunnel feature enabled:

    ```shell
    kubectl kcp workload sync kind --resources=poddisruptionbudgets.policy,horizontalpodautoscalers.autoscaling,services,endpoints,pods --syncer-image kind.local/syncer --feature-gates KCPSyncerTunnel=true -o kind-syncer.yaml
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

1. Finally, bind compute (ie. Deployments, Services and Ingresses):

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
    kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.7.0/serving-crds.yaml
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
    kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.7.0/serving-core.yaml
    ```

   > Note: ignore the last two errors `no matches for kind "HorizontalPodAutoscaler" in version "autoscaling/v2beta2`

1. Annotate `activator-service` indicating kcp to synchronize derived endpoints to kcp:

   ```shell
   kubectl annotate -n knative-serving svc activator-service workload.kcp.io/upsync-derived-resources=endpoints
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
   kubectl apply -f https://github.com/knative/net-kourier/releases/download/knative-v1.7.0/kourier.yaml
   ```

1. Annotate `kourier-internal` indicating kcp to synchronize derived endpoints to kcp:

   ```shell
   kubectl annotate -n kourier-system svc kourier-internal workload.kcp.io/upsync-derived-resources=endpoints
   ```

7. Since KCP does not support service-based admission controllers the config map validating
   webhook needs to be deleted:

   ```shell
   kubectl delete validatingwebhookconfigurations.admissionregistration.k8s.io --all
   kubectl delete mutatingwebhookconfigurations.admissionregistration.k8s.io  --all
   ```

8. Patch the network configmap:

   ```shell
   kubectl patch configmap/config-network \
           --namespace knative-serving \
           --type merge \
           --patch '{"data":{"ingress-class":"kourier.ingress.networking.knative.dev"}}'
   ```

9. Verify kourier is up and running:

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

## Deploying your first Knative service

In the KCP terminal, deploy the hello world app:

```shell
cat <<EOF | kubectl apply -f -
apiVersion: serving.knative.dev/v1
kind: Service
metadata:
  name: hello
spec:
  template:
    metadata:
      annotations:
        autoscaling.knative.dev/class: kpa.autoscaling.knative.dev
        workload.kcp.io/upsync-derived-resources: endpoints
    spec:
      containers:
      - env:
        - name: TARGET
          value: World
        image: gcr.io/knative-samples/helloworld-go
        ports:
        - containerPort: 8080
          protocol: TCP
EOF
```

Note that:
  - `autoscaling.knative.dev/class: kpa.autoscaling.knative.dev` must be explicitly set because the defaulting webhook was disabled
  -  `workload.kcp.io/upsync-derived-resources: endpoints` must be set so that Knative can detect the service is ready

## TODOs

TODOs:
- Magic DNS/Real DNS
- Eventing
- HPA
- Deleting service
