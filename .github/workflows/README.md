# CI/CD Pipeline

GitHub Actions workflow untuk build dan deploy ke Google Kubernetes Engine (GKE).

## Pipeline Overview

```
┌─────────┐     ┌─────────┐     ┌────────────────┐
│  Test   │ ──▶ │  Build  │ ──▶ │     Deploy     │
└─────────┘     └─────────┘     └────────────────┘
                                  │
                                  ├── develop (auto)
                                  └── production (manual)
```

## Triggers

| Event | Branch | Action |
|-------|--------|--------|
| Push | `develop` | Test → Build → Deploy to Develop |
| Push | `main` | Test → Build (no deploy) |
| Pull Request | `main` | Test only |
| Manual | - | Deploy to Production |

## Required Variables

Tambahkan di **Settings → Secrets and variables → Actions → Variables**

| Variable | Description | Example |
|----------|-------------|---------|
| `GCP_PROJECT_ID` | Google Cloud Project ID | `free-tier-project-488416` |
| `WIF_PROVIDER` | Workload Identity Provider | `projects/PROJECT_NUMBER/locations/global/workloadIdentityPools/github-pool/providers/github-provider` |
| `GCP_SERVICE_ACCOUNT` | Service Account email | `github-actions@PROJECT_ID.iam.gserviceaccount.com` |
| `GKE_CLUSTER_DEV` | Nama cluster development | `dev-cluster` |
| `GKE_ZONE_DEV` | Zone cluster development | `asia-southeast2-a` |
| `GKE_CLUSTER_PROD` | Nama cluster production | `prod-cluster` |
| `GKE_ZONE_PROD` | Zone cluster production | `asia-southeast2-a` |

> **Note:** Tidak perlu secret! Menggunakan Workload Identity Federation (lebih aman).

## Setup Workload Identity Federation

### 1. Create Service Account

```bash
PROJECT_ID="your-project-id"

gcloud iam service-accounts create github-actions \
  --display-name="GitHub Actions CI/CD"
```

### 2. Grant Permissions

```bash
# Artifact Registry - push images
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:github-actions@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role="roles/artifactregistry.writer"

# GKE - deploy to cluster
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:github-actions@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role="roles/container.developer"
```

### 3. Create Workload Identity Pool

```bash
gcloud iam workload-identity-pools create "github-pool" \
  --location="global" \
  --display-name="GitHub Actions Pool"
```

### 4. Create OIDC Provider

```bash
gcloud iam workload-identity-pools providers create-oidc "github-provider" \
  --location="global" \
  --workload-identity-pool="github-pool" \
  --display-name="GitHub Provider" \
  --attribute-mapping="google.subject=assertion.sub,attribute.actor=assertion.actor,attribute.repository=assertion.repository" \
  --attribute-condition="assertion.repository=='YOUR_GITHUB_USERNAME/YOUR_REPO'" \
  --issuer-uri="https://token.actions.githubusercontent.com"
```

### 5. Allow GitHub to impersonate Service Account

```bash
PROJECT_NUMBER=$(gcloud projects describe $PROJECT_ID --format='value(projectNumber)')

gcloud iam service-accounts add-iam-policy-binding \
  github-actions@${PROJECT_ID}.iam.gserviceaccount.com \
  --role="roles/iam.workloadIdentityUser" \
  --member="principalSet://iam.googleapis.com/projects/${PROJECT_NUMBER}/locations/global/workloadIdentityPools/github-pool/attribute.repository/YOUR_GITHUB_USERNAME/YOUR_REPO"
```

### 6. Get WIF Provider value

```bash
echo "projects/${PROJECT_NUMBER}/locations/global/workloadIdentityPools/github-pool/providers/github-provider"
```

Copy output ini ke variable `WIF_PROVIDER` di GitHub.

## Manual Deploy to Production

1. Go to **Actions** tab
2. Select **CI/CD Pipeline**
3. Click **Run workflow**
4. Select `production` environment
5. Click **Run workflow**

## Deployment Files

- `deploy/k8s/deployment.dev.yaml` - Development config
- `deploy/k8s/deployment.prod.yaml` - Production config
