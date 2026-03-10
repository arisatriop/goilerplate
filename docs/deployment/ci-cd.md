# CI/CD Pipeline Guide

GitHub Actions workflow for automated build, test, and deploy to Google Kubernetes Engine (GKE).

---

## 🔄 Pipeline Overview

```
┌─────────┐     ┌─────────┐     ┌────────────────┐
│  Test   │ ──▶ │  Build  │ ──▶ │     Deploy     │
└─────────┘     └─────────┘     └────────────────┘
                                  │
                                  ├── develop (auto)
                                  ├── staging (auto)
                                  └── production (manual)
```

---

## 🎯 Triggers & Behavior

| Event      | Branch     | Action                                    |
|------------|-----------|-------------------------------------------|
| Push       | `develop` | Test → Build → Deploy to Development      |
| Push       | `staging` | Test → Build → Deploy to Staging          |
| Push       | `main`    | Test → Build (⚠️ Manual Deploy)           |
| Manual     | `main`    | Deploy to Production                      |

---

## 📝 Required GitHub Variables

Setup di **Settings → Secrets and variables → Actions → Variables**

| Variable            | Description                          | Example |
|---------------------|--------------------------------------|---------|
| `GCP_PROJECT_ID`    | Google Cloud Project ID              | `free-tier-project-488416` |
| `WIF_PROVIDER`      | Workload Identity Provider           | `projects/PROJECT_NUMBER/locations/global/workloadIdentityPools/github-pool/providers/github-provider` |
| `GCP_SERVICE_ACCOUNT`| Service Account email                | `github-actions@PROJECT_ID.iam.gserviceaccount.com` |
| `GKE_CLUSTER_DEV`   | Development cluster name             | `dev-cluster` |
| `GKE_CLUSTER_STAG`  | Staging cluster name                 | `stag-cluster` |
| `GKE_CLUSTER_PROD`  | Production cluster name              | `prod-cluster` |
| `GKE_ZONE`          | Cluster zone (all environments)      | `asia-southeast2-a` |

> **Note:** No need to use Secrets for authentication. Pipeline uses **Workload Identity Federation** (more secure).

---

## 🔐 Setup Workload Identity Federation (WIF)

Workload Identity Federation is a secure way to authenticate GitHub Actions to GCP without storing service account keys.

### 1️⃣ Create Service Account

```bash
PROJECT_ID="your-project-id"

gcloud iam service-accounts create github-actions \
  --display-name="GitHub Actions CI/CD"
```

### 2️⃣ Grant Permissions

**Artifact Registry (for pushing Docker image):**
```bash
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:github-actions@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role="roles/artifactregistry.writer"
```

**GKE (for deploying to cluster):**
```bash
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:github-actions@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role="roles/container.developer"
```

### 3️⃣ Create Workload Identity Pool

```bash
gcloud iam workload-identity-pools create "github-pool" \
  --location="global" \
  --display-name="GitHub Actions Pool"
```

### 4️⃣ Create OIDC Provider

```bash
gcloud iam workload-identity-pools providers create-oidc "github-provider" \
  --location="global" \
  --workload-identity-pool="github-pool" \
  --display-name="GitHub Provider" \
  --attribute-mapping="google.subject=assertion.sub,attribute.actor=assertion.actor,attribute.repository=assertion.repository" \
  --attribute-condition="assertion.repository=='YOUR_GITHUB_USERNAME/YOUR_REPO'" \
  --issuer-uri="https://token.actions.githubusercontent.com"
```

### 5️⃣ Allow GitHub to Impersonate Service Account

```bash
PROJECT_NUMBER=$(gcloud projects describe $PROJECT_ID --format='value(projectNumber)')

gcloud iam service-accounts add-iam-policy-binding \
  github-actions@${PROJECT_ID}.iam.gserviceaccount.com \
  --role="roles/iam.workloadIdentityUser" \
  --member="principalSet://iam.googleapis.com/projects/${PROJECT_NUMBER}/locations/global/workloadIdentityPools/github-pool/attribute.repository/YOUR_GITHUB_USERNAME/YOUR_REPO"
```

### 6️⃣ Get WIF Provider Value

```bash
echo "projects/${PROJECT_NUMBER}/locations/global/workloadIdentityPools/github-pool/providers/github-provider"
```

Copy this output to GitHub variable `WIF_PROVIDER`.

---

## 🚀 Manual Deploy to Production

### Via GitHub Actions UI

1. Go to **Actions** tab
2. Select **CI/CD Pipeline** workflow
3. Click **Run workflow**
4. Select `main` branch
5. Click **Run workflow**

Workflow will:
- Run tests
- Build image
- Push to registry
- Deploy to production cluster

---

## 📂 Deployment Files

Project includes pre-configured deployment files:

```
deploy/k8s/
├── deployment.dev.yaml  # Development config
├── deployment.stag.yaml # Staging config
└── deployment.prod.yaml # Production config
```

---

## 📊 Pipeline Jobs

### 1. Test Job
- Run `go test ./...`
- Run linting (jika ada)
- Validate Go code

### 2. Build Job
- Build Docker image
- Tag with git commit SHA
- Push to Artifact Registry

### 3. Deploy Job (conditional)
- Deploy to GKE cluster (based on branch)
- Apply ConfigMap & Secret
- Update deployment with new image

---

## 🔍 Monitoring Pipeline

### View Workflow Runs
```bash
gh run list
gh run view <run-id>
```

### View Logs
```bash
gh run view <run-id> --log
```

### Check Deployment Status
```bash
kubectl get deployment -n <namespace>
kubectl rollout status deployment/goilerplate -n <namespace>
kubectl logs -f deployment/goilerplate -n <namespace>
```

---

## 🐛 Troubleshooting

### Auth Failed
- Verify WIF_PROVIDER variable is correct
- Check service account permissions
- Verify GitHub repo name in OIDC provider condition

### Image Push Failed
- Check Artifact Registry permissions
- Verify registry URL in workflow
- Check Docker image naming convention

### Deployment Failed
- Check ConfigMap & Secret exist
- Verify cluster zone is correct
- Check resource limits & requests
- View pod logs: `kubectl logs <pod-name>`

---

## 🔐 Best Practices

✅ **DO:**
- Use Workload Identity Federation (no need to store keys)
- Tag images with semantic versioning
- Test in staging before production
- Monitor deployments with logs & metrics
- Rotate service account permissions regularly

❌ **DON'T:**
- Store GCP service account key in GitHub
- Use `latest` tag in production
- Auto-deploy to production (always manual)
- Skip testing before deploy
- Mix secrets in ConfigMap

---

## 🔗 Related

- [Kubernetes Deployment](./kubernetes.md) - Deployment configuration
- [Configuration Guide](./configuration.md) - Environment setup
- [GitHub Actions Workflow](.github/workflows)
