# Exporting CRDs

This guide goes through the steps to export CRDs as APIs from one worskspace
and bind them from other workspaces.

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

- Go to your home workspace:
  
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

- Verify it has been correctly imported:

```shell
kubectl api-resources | grep stable.example.com
```

```shell
crontabs  ct stable.example.com/v1 true CronTab
```

- Create an CR:

```shell
kubectl apply -f mycrontab.yaml
```

```shell
crontab.stable.example.com/my-new-cron-object created
```


