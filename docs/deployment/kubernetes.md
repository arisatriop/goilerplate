# Kubernetes Deployment Guide

Guide to deploy Goilerplate to Google Kubernetes Engine (GKE) or other Kubernetes clusters.

---

## 📋 Prerequisites

- Kubernetes cluster (GKE, EKS, AKS, or local)
- `kubectl` CLI installed and configured
- Docker image already pushed to container registry
- ConfigMap and Secret already set up

---

## 🔧 Creating ConfigMap & Secret

Before deployment, set up ConfigMap for non-sensitive values and Secret for sensitive values.

### ConfigMap

**Create ConfigMap from file:**

```bash
kubectl create configmap goilerplate-config -n <namespace> \
  --from-file=config.yaml=./config/config.example.yaml \
  --dry-run=client -o yaml | kubectl apply -f -
```

**Or create YAML file first:**

```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: goilerplate-config
  namespace: default
data:
  config.yaml: |
    app:
      env: production
      name: Goilerplate
      version: 1.0.0
    server:
      host: 0.0.0.0
      port: 3000
    db:
      host: your-db-host
      port: 5432
    redis:
      enabled: true
      host: your-redis-host:6379
```

Apply dengan:
```bash
kubectl apply -f configmap.yaml
```

---

### Secret (from .env file)

**Create Secret from env file:**

```bash
kubectl create secret generic goilerplate-secret -n <namespace> \
  --from-env-file=./config/.env \
  --dry-run=client -o yaml | kubectl apply -f -
```

---

### Secret (from literal values)

If there's no `.env` file, create from literal values:

```bash
kubectl create secret generic goilerplate-secret -n <namespace> \
  --from-literal=DB_HOST=your-db-host \
  --from-literal=DB_PORT=5432 \
  --from-literal=DB_NAME=goilerplate \
  --from-literal=DB_USERNAME=postgres \
  --from-literal=DB_PASSWORD=your-secret-password \
  --from-literal=REDIS_HOST=your-redis-host:6379 \
  --from-literal=JWT_ACCESS_SECRET=... \
  --from-literal=JWT_REFRESH_SECRET=... \
  --dry-run=client -o yaml | kubectl apply -f -
```

The access and refresh tokens are signed with **separate** secrets, so that an attacker who
obtains one cannot mint the other. There is no combined `JWT_SECRET_KEY`. Generate each with
`openssl rand -base64 48`; startup validation rejects anything under 32 bytes, and with
`app.env=production` it also rejects the example secrets that ship in `config.example.yaml`.

---

## 📦 Deployment Files

Project includes deployment files for various environments:

- `deploy/k8s/deployment.dev.yaml` - Development config
- `deploy/k8s/deployment.stag.yaml` - Staging config
- `deploy/k8s/deployment.prod.yaml` - Production config

---

## 🚪 Ingress: never expose `/internal`

The app serves three route groups on one port:

| Prefix | Who may reach it | Auth |
|---|---|---|
| `/api` | the public internet | Bearer JWT (or none, for login and register) |
| `/partner` | named partners | `x-api-key` |
| `/livez`, `/readyz` | probes, the gateway | none |
| `/internal` | **other pods only** | none by default |

`/internal` has no user authentication. It is reachable by anything that can open a connection
to the pod, so what keeps it private is the Ingress, and nothing else.

### The rule

**Forward an explicit allowlist. Never a catch-all `/`.**

```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: goilerplate
spec:
  rules:
    - host: api.example.com
      http:
        paths:
          # ✅ one rule per public prefix
          - path: /api
            pathType: Prefix
            backend: { service: { name: goilerplate, port: { number: 3000 } } }
          - path: /partner
            pathType: Prefix
            backend: { service: { name: goilerplate, port: { number: 3000 } } }
          - path: /livez
            pathType: Exact
            backend: { service: { name: goilerplate, port: { number: 3000 } } }
          - path: /readyz
            pathType: Exact
            backend: { service: { name: goilerplate, port: { number: 3000 } } }

          # ❌ never this — it publishes /internal along with everything else
          # - path: /
          #   pathType: Prefix
```

Where the controller supports it, deny `/internal` explicitly as well, so a later edit that adds
a catch-all does not silently undo this:

```yaml
metadata:
  annotations:
    nginx.ingress.kubernetes.io/server-snippet: |
      location ~* ^/internal { deny all; return 404; }
```

### Verify it

From **outside** the cluster. The request must not reach the app:

```bash
curl -i https://api.example.com/internal/bars
# want: 404 from the ingress controller, or a connection that never reaches a pod
# BAD:  200, or any response carrying X-Request-Id (that header means the app answered)
```

Then confirm the route does work from inside, so you know the test above proved something:

```bash
kubectl run curl --rm -it --image=curlimages/curl --restart=Never -n <namespace> -- \
  curl -s -o /dev/null -w '%{http_code}\n' http://goilerplate:3000/internal/bars
# want: 200
```

`X-Request-Id` is the tell. Every response the app produces carries it; an ingress 404 does not.

### Second lock: `internal_auth`

An Ingress is one config file, often owned by whoever runs the cluster rather than by whoever
owns this service. When that is the case — or the config changes often enough that one day it
will be wrong — turn on the shared secret:

```yaml
internal_auth:
  mode: shared_secret
  secret: <a random value of at least 32 bytes>
```

Callers then send it:

```
X-Internal-Secret: <the secret>
```

The header is compared in constant time and is redacted from logs. In production the app logs a
startup warning while `mode` is `none`, so the reliance on the gateway is at least stated out
loud rather than assumed.

This is a *second* lock, not a replacement for the allowlist: the secret is shared by every
caller and travels on every request, so it proves "something in the cluster", not which service.
`X-Service-Name` is recorded alongside it for attribution in logs and is an unverified claim —
never use it for authorization.

---

## 🚀 Deploy to Kubernetes

### 1. Update Image Reference

Edit deployment file and update container image:

```yaml
containers:
  - name: goilerplate
    image: your-registry/goilerplate:latest  # Update this
    ports:
      - containerPort: 3000
    envFrom:
      - secretRef:
          name: goilerplate-secret
    volumeMounts:
      - name: config
        mountPath: /app/config
  volumes:
    - name: config
      configMap:
        name: goilerplate-config
```

### 2. Apply Deployment

**Development:**
```bash
kubectl apply -f deploy/k8s/deployment.dev.yaml -n development
```

**Staging:**
```bash
kubectl apply -f deploy/k8s/deployment.stag.yaml -n staging
```

**Production:**
```bash
kubectl apply -f deploy/k8s/deployment.prod.yaml -n production
```

---

## ✅ Verify Deployment

### Check Pod Status
```bash
kubectl get pods -n <namespace>
kubectl describe pod <pod-name> -n <namespace>
kubectl logs <pod-name> -n <namespace>
```

### Port Forward (Testing)
```bash
kubectl port-forward svc/goilerplate 3000:3000 -n <namespace>
curl http://localhost:3000/readyz
```

### Check ConfigMap & Secret
```bash
kubectl get configmap -n <namespace>
kubectl get secret -n <namespace>
kubectl describe configmap goilerplate-config -n <namespace>
```

---

## 🔄 Updating Deployment

### Update Image
```bash
kubectl set image deployment/goilerplate \
  goilerplate=your-registry/goilerplate:v1.1.0 \
  -n <namespace>
```

### Update ConfigMap
Edit and reapply:
```bash
kubectl apply -f configmap.yaml
# Restart pods to load new config
kubectl rollout restart deployment/goilerplate -n <namespace>
```

### Rollback
```bash
kubectl rollout history deployment/goilerplate -n <namespace>
kubectl rollout undo deployment/goilerplate -n <namespace>
```

---

## 📊 Environment-Specific Configs

### Development
- Replicas: 1
- Resource requests: Low
- Image pull policy: Always (for testing)

### Staging
- Replicas: 2
- Resource requests: Medium
- Image pull policy: IfNotPresent

### Production
- Replicas: 3+
- Resource requests: High
- Image pull policy: IfNotPresent
- Health checks: Enabled
- Rolling updates: Configured

---

## 🔐 Best Practices

✅ **DO:**
- Route an explicit allowlist at the Ingress; never a catch-all `/` (see above)
- Use separate namespaces for each environment
- Store secrets in Secret, not in ConfigMap
- Use health checks (liveness & readiness probes)
- Set resource limits and requests
- Use rolling updates for zero downtime
- Monitor logs and metrics

❌ **DON'T:**
- Expose `/internal` through the public Ingress
- Store secrets in ConfigMap
- Hardcode values in YAML
- Use `latest` tag in production
- Skip health checks
- Deploy without rolling strategy

---

## 🔗 Related

- [Configuration Guide](./configuration.md) - Setup environment variables
- [CI/CD Pipeline](./ci-cd.md) - Automated deployment with GitHub Actions
- [Main Deployment Directory](../../deploy/k8s)
