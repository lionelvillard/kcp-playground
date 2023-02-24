- In one terminal, start kcp (`kcp start`)
- Open a new terminal
- Run `export KUBECONFIG=$(pwd)/.kcp/admin.kubeconfig`
- Create provider ws `kubectl kcp ws create provider --enter`
- Create an APIExport claiming namespaces:

```shell
cat <<EOF | kubectl apply -f -
apiVersion: apis.kcp.io/v1alpha1
kind: APIExport
metadata:
  name: example
spec:
  latestResourceSchemas: []
  permissionClaims:
    - resource: namespaces
      all: true
EOF
```

- Go to your ws `kubectl ws`
- Bind the APIExport example `kubectl kcp bind apiexport root:provider:example`
- Get your workspace id `export MYWS=$(kubectl get namespace default -ojsonpath="{.metadata.annotations['kcp\.io/cluster']}")`
- Go back to the provider ws `kubectl ws root:provider`
- Allow provider admin to access views

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

- Get the APIExport view URL `export VWURL=$(kubectl get apiexport example -ojsonpath='{.status.virtualWorkspaces[0].url}')`
- Get the provider tokenAPIExport view URL `export PROVIDER_TOKEN=$(kubectl get secret default-token-hl48b -ojsonpath='{.data.token}' | base64 -D)`
- Get the list of your workspace namespace via the view:

```
kubectl --server=$VWURL/clusters/* --token=$PROVIDER_TOKEN get namespaces
No resources found
```

- Create a new namespace:

```
kubectl --server=$VWURL/clusters/$MYWS --token=$PROVIDER_TOKEN create namespace boo
namespace/boo created
```
