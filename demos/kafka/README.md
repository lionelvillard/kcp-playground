# Self-Service Apache Kafka

This is a demo showing how to offer Apache Kafka as a service using kcp.

## Goal

The goal is to trivialize setting up environments needed to run applications
depending on external services. Suppose your application depends on Apache Kafka,
this one-liner creates a workspace containing the necessary objects to access an Apache Kafka cluster
ready for your application use:

```shell
kubectl kcp ws create my-kafka-app --type root:platform:kafka:local
```

Let's see how that's done using kcp API export/binding and configuration management capabilities.

## Prerequisites

Make sure you have configured `KUBECONFIG` to access kcp. See [kcp quickstart](https://docs.kcp.io/kcp/main) for more detailed instructions.

## Organization Workspaces

Suppose your organization has a platform engineering team using kcp as the internal developer
platform backend to offer self-service services such as Apache Kafka. This team provides all
services from the `platform` workspace:

- Create the `platform` workspace where all resources related to the organization platform will be stored:

  ```shell
  kubectl ws create platform --enter
  ```

  ```shell
  Workspace "platform" (type root:organization) created. Waiting for it to be ready...
  Workspace "platform" (type root:organization) is ready to use.
  Current workspace is "root:platform" (type root:organization).
  ```

Application developers can create workspaces under the application workspaces they have access to.

- Create the `applications` workspace containing all applications in your organization:

  ```shell
  kubectl ws root && kubectl ws create applications --enter
  ```

  ```shell
  Workspace "applications" (type root:organization) created. Waiting for it to be ready...
  Workspace "applications" (type root:organization) is ready to use.
  Current workspace is "root:applications" (type root:organization).
  ```

## Setting up the platform offerings

### Prepare the Apache Kafka Service workspace types

- Create the `kafka` workspace to store all Kafka related resources:

  ```shell
  kubectl ws root:platform && kubectl ws create kafka --enter
  ```

  ```shell
  Workspace "kafka" (type root:universal) created. Waiting for it to be ready...
  Workspace "kafka" (type root:universal) is ready to use.
  Current workspace is "root:platform:kafka" (type root:universal).
  ```

- Allow admin user to access exported content:

  ```shell
  cat <<EOF | kubectl apply -f -
  apiVersion: rbac.authorization.k8s.io/v1
  kind: ClusterRole
  metadata:
    name: default
  rules:
    - apiGroups:
        - apis.kcp.io
      resources:
        - apiexports/content
      verbs:
        - "*"
  ---
  apiVersion: rbac.authorization.k8s.io/v1
  kind: ClusterRoleBinding
  metadata:
    name: default
  roleRef:
    apiGroup: rbac.authorization.k8s.io
    kind: ClusterRole
    name: default
  subjects:
    - apiGroup:
      kind: ServiceAccount
      name: default
      namespace: default
  EOF
  ```

### Apache Kafka Environments

#### Local Development

- Create the APIExport object giving access to a single Kafka cluster. No APIs are exposed, only a
  permission request to create/update a secret containing the Kafka cluster credential:

  ```shell
  kubectl apply -f platform/envs/apiexport-local.yaml
  ```

- Create the APILifecycle object registering the webhook responsible for creating the aforementioned secret:

  ```shell
  kubectl apply -f platform/envs/apilifecycle-local.yaml
  ```

- Finally, create the WorkspaceType which automatically binds to the local Kafka environment:

  ```shell
  kubectl apply -f platform/envs/workspacetype-local.yaml
  ```

#### Running webhooks

In a terminal, start the webhook server:

```shell
go run platform/webhook/main.go
```

## Writing your Kafka-based application

Create a workspace hosting all assets for your application:

```shell
kubectl ws root:applications && kubectl ws create hello-kafka --type root:platform:kafka:local --enter
```

You can verify that a secret containing the service credentials have been created in the default namespace:

```shell
kubectl get secret kafka-credentials -n default
```

```shell
NAME            TYPE     DATA   AGE
kafka-credentials   Opaque   1      2m22s
```

This secret can be imported by your application.


## In Progress

- Real Kafka credential
- More environment
- Integrate crossplane
- gitops
- observability
- knative
