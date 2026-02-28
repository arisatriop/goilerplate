# Creating ConfigMap and Secret
You need to ConfigMap or Secret before deploying the application. Creating both is recommended: non-sensitive value store in ConfigMap, and override with Secret for sensitive value. You can also use .env file to store all the values.

## ConfigMap
To create a ConfigMap from a file:

```sh
kubectl create configmap goilerplate-config -n <namespace> \
  --from-file=config.yaml=./config/config.example.yaml \
  --dry-run=client -o yaml | kubectl apply -f -
```

## Secret
To create a Secret from a file:

```sh
kubectl create secret generic goilerplate-secret -n <namespace> \
  --from-env-file=./config/.env \
  --dry-run=client -o yaml | kubectl apply -f -
```

Or from literal values:

```sh
kubectl create secret generic goilerplate-secret -n <namespace> \
	--from-literal=key1=value1 \
	--from-literal=key2=value2 \
    --dry-run=client -o yaml | kubectl apply -f -
```
