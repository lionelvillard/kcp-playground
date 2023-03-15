# Lifecycle WebHooks

This guide shows how to use kcp API lifecyle webhooks

DISCLAIMER: this is not an official kcp feature.

- start kcp

    ```shell
    kcp start
    ```

- Create a workspace:

  ```shell
  kubectl kcp ws create provider --enter
  ```

- Convert the example CRD to an APIResourceSchema and apply it:

  ```shell
  kubectl kcp crd snapshot --filename example-crd.yaml --prefix v1 | kubectl apply -f -
  ```

- Then export the API:

  ```shell
  kubectl apply -f example-apiexport.yaml
  ```

- Add lifecycle hooks:

```shell
kubectl apply -f example-apilifecycle.yaml
```

- Allow admin to access exported content:

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

- In a new terminal, start the webhook:

  ```shell
  go run webhook/main.go
  ```

- In the kcp terminal, go to your home workspace:

```shell
kubectl kcp ws
```

- Then import the API:

```shell
kubectl kcp bind apiexport root:provider:example
```

```shell
apibinding example created. Waiting to successfully bind ...
example created and bound.
```

- Verify the secret returned by the webhook has been created:

```shell
kubectl get secret from-provider -oyaml
```

```shell
apiVersion: v1
data:
  avery: c2VjcmV0
kind: Secret
metadata:
  annotations:
    kcp.io/cluster: kvdk2spgmbix
  creationTimestamp: "2023-03-13T19:26:29Z"
  name: from-provider
  namespace: default
  resourceVersion: "800"
  uid: 5118ccac-33f7-453e-83cd-0cb6f711be32
type: Opaque
```
