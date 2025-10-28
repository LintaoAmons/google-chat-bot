# GCP Cloud Run Deployment Setup

This guide will help you set up automatic deployment to GCP Cloud Run using GitHub Actions.

## Prerequisites

- A Google Cloud Platform account
- Docker installed locally
- Admin access to your GCP project
- Admin access to your GitHub repository

## Step 0: Initialize gcloud CLI via Docker

All gcloud commands in this guide use Docker, so you don't need to install the gcloud CLI locally.

### 0.1 Authenticate and Configure gcloud

Run this command to authenticate and set your project (replace `YOUR_PROJECT_ID` with your actual GCP project ID):

```bash
docker run -ti --name gcloud-config gcr.io/google.com/cloudsdktool/google-cloud-cli:stable \
/bin/bash -c 'gcloud auth login && gcloud config set project YOUR_PROJECT_ID'
```

This will:
1. Open a browser window for you to authenticate with your Google account
2. Create a persistent Docker container named `gcloud-config` that stores your credentials
3. Set your default project

**Note:** The `gcloud-config` container will be reused by all subsequent commands via `--volumes-from gcloud-config`.

### 0.2 Find Your Project ID

**IMPORTANT:** GCP has two different project identifiers:
- **Project ID** (string): e.g., `my-chat-bot-project` ← **Use this in all commands**
- **Project Number** (numeric): e.g., `729450032813` ← **Don't use this**

List all your projects to find the **Project ID**:

```bash
docker run --rm --volumes-from gcloud-config gcr.io/google.com/cloudsdktool/google-cloud-cli:stable \
gcloud projects list
```

Output example:
```
PROJECT_ID              NAME                PROJECT_NUMBER
my-chat-bot-project     Google Chat Bot     729450032813
↑ Use this one                              ↑ NOT this
```

Use the value from the **PROJECT_ID** column (first column) in all subsequent commands.

### 0.3 Verify Current Project (Optional)

```bash
docker run --rm --volumes-from gcloud-config gcr.io/google.com/cloudsdktool/google-cloud-cli:stable \
gcloud config get-value project
```

This should display your project ID (the string, not the number).

## Step 1: Set Up GCP Project

### 1.1 Create a New Project (Optional)

If you need to create a new project:

```bash
docker run --rm --volumes-from gcloud-config gcr.io/google.com/cloudsdktool/google-cloud-cli:stable \
gcloud projects create YOUR_PROJECT_ID --name="Google Chat Bot"
```

Then update your config to use the new project:

```bash
docker run --rm --volumes-from gcloud-config gcr.io/google.com/cloudsdktool/google-cloud-cli:stable \
gcloud config set project YOUR_PROJECT_ID
```

### 1.2 Enable Required APIs

```bash
# IMPORTANT: Replace YOUR_PROJECT_ID with your actual GCP project ID throughout these commands

# Enable Cloud Run API
docker run --rm --volumes-from gcloud-config gcr.io/google.com/cloudsdktool/google-cloud-cli:stable \
gcloud services enable run.googleapis.com --project=YOUR_PROJECT_ID

# Enable IAM Credentials API (for Workload Identity Federation)
docker run --rm --volumes-from gcloud-config gcr.io/google.com/cloudsdktool/google-cloud-cli:stable \
gcloud services enable iamcredentials.googleapis.com --project=YOUR_PROJECT_ID

# Enable Cloud Resource Manager API
docker run --rm --volumes-from gcloud-config gcr.io/google.com/cloudsdktool/google-cloud-cli:stable \
gcloud services enable cloudresourcemanager.googleapis.com --project=YOUR_PROJECT_ID
```

## Step 2: Set Up Workload Identity Federation

Workload Identity Federation allows GitHub Actions to authenticate to GCP without storing service account keys.

### 2.1 Create a Service Account

```bash
# Create service account for GitHub Actions
docker run --rm --volumes-from gcloud-config gcr.io/google.com/cloudsdktool/google-cloud-cli:stable \
gcloud iam service-accounts create github-actions-cloudrun \
  --display-name="GitHub Actions Cloud Run Deployer" \
  --project=YOUR_PROJECT_ID

# Get your project number (you'll need this later)
docker run --rm --volumes-from gcloud-config gcr.io/google.com/cloudsdktool/google-cloud-cli:stable \
gcloud projects describe YOUR_PROJECT_ID --format='value(projectNumber)'

# Grant Cloud Run Admin role
docker run --rm --volumes-from gcloud-config gcr.io/google.com/cloudsdktool/google-cloud-cli:stable \
gcloud projects add-iam-policy-binding YOUR_PROJECT_ID \
  --member="serviceAccount:github-actions-cloudrun@YOUR_PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/run.admin"

# Grant Service Account User role (required to deploy as another service account)
docker run --rm --volumes-from gcloud-config gcr.io/google.com/cloudsdktool/google-cloud-cli:stable \
gcloud projects add-iam-policy-binding YOUR_PROJECT_ID \
  --member="serviceAccount:github-actions-cloudrun@YOUR_PROJECT_ID.iam.gserviceaccount.com" \
  --role="roles/iam.serviceAccountUser"
```

### 2.2 Create Workload Identity Pool
> TODO: What's this?

```bash
# Create the workload identity pool
docker run --rm --volumes-from gcloud-config gcr.io/google.com/cloudsdktool/google-cloud-cli:stable \
gcloud iam workload-identity-pools create "github-actions-pool" \
  --project="YOUR_PROJECT_ID" \
  --location="global" \
  --display-name="GitHub Actions Pool"

# Create the workload identity provider
docker run --rm --volumes-from gcloud-config gcr.io/google.com/cloudsdktool/google-cloud-cli:stable \
gcloud iam workload-identity-pools providers create-oidc "github-provider" \
  --project="YOUR_PROJECT_ID" \
  --location="global" \
  --workload-identity-pool="github-actions-pool" \
  --display-name="GitHub Provider" \
  --attribute-mapping="google.subject=assertion.sub,attribute.actor=assertion.actor,attribute.repository=assertion.repository,attribute.repository_owner=assertion.repository_owner" \
  --attribute-condition="assertion.repository_owner == 'YOUR_GITHUB_USERNAME'" \
  --issuer-uri="https://token.actions.githubusercontent.com"

  # Delete the entire workload identity pool (this deletes all providers inside it too), then you can recreate
  docker run --rm --volumes-from gcloud-config gcr.io/google.com/cloudsdktool/google-cloud-cli:stable \
  gcloud iam workload-identity-pools delete "github-actions-pool" \
    --project="YOUR_PROJECT_ID" \
    --location="global" \
    --quiet
```

> TODO: bookmark, left here

**Important:** Replace `YOUR_GITHUB_USERNAME` with your GitHub username or organization name.

### 2.3 Allow GitHub Actions to Impersonate the Service Account

```bash
# Get the workload identity pool ID (save this output, you'll need it for the next command)
docker run --rm --volumes-from gcloud-config gcr.io/google.com/cloudsdktool/google-cloud-cli:stable \
gcloud iam workload-identity-pools describe "github-actions-pool" \
  --project="YOUR_PROJECT_ID" \
  --location="global" \
  --format="value(name)"

# Grant the service account permission to be impersonated
# Replace WORKLOAD_IDENTITY_POOL_ID with the full path from the previous command
# It looks like: projects/123456789/locations/global/workloadIdentityPools/github-actions-pool
docker run --rm --volumes-from gcloud-config gcr.io/google.com/cloudsdktool/google-cloud-cli:stable \
gcloud iam service-accounts add-iam-policy-binding \
  "github-actions-cloudrun@YOUR_PROJECT_ID.iam.gserviceaccount.com" \
  --project="YOUR_PROJECT_ID" \
  --role="roles/iam.workloadIdentityUser" \
  --member="principalSet://iam.googleapis.com/WORKLOAD_IDENTITY_POOL_ID/attribute.repository/YOUR_GITHUB_USERNAME/google-chat-bot"
```

**Important:** Replace `YOUR_GITHUB_USERNAME/google-chat-bot` with your actual repository path.

### 2.4 Get the Workload Identity Provider Resource Name

```bash
docker run --rm --volumes-from gcloud-config gcr.io/google.com/cloudsdktool/google-cloud-cli:stable \
gcloud iam workload-identity-pools providers describe "github-provider" \
  --project="YOUR_PROJECT_ID" \
  --location="global" \
  --workload-identity-pool="github-actions-pool" \
  --format="value(name)"
```

This will output something like:
```
projects/123456789/locations/global/workloadIdentityPools/github-actions-pool/providers/github-provider
```

**Save this value** - you'll need it for the `GCP_WORKLOAD_IDENTITY_PROVIDER` GitHub secret.

## Step 3: Configure GitHub Secrets

Go to your GitHub repository settings and add the following secrets:

### Required Secrets

1. **GCP_PROJECT_ID**
   - Your GCP project ID (e.g., `my-project-123`)

2. **GCP_WORKLOAD_IDENTITY_PROVIDER**
   - The full resource name from Step 2.4 above
   - Format: `projects/PROJECT_NUMBER/locations/global/workloadIdentityPools/github-actions-pool/providers/github-provider`

3. **GCP_SERVICE_ACCOUNT**
   - Your service account email
   - Format: `github-actions-cloudrun@YOUR_PROJECT_ID.iam.gserviceaccount.com`

4. **GOOGLE_CHAT_WEBHOOK_URL**
   - Your Google Chat webhook URL (this will be set as an environment variable in Cloud Run)
   - Format: `https://chat.googleapis.com/v1/spaces/...`

### Existing Secrets (Already Configured)

These should already exist for your Docker Hub workflow:
- `DOCKERHUB_USERNAME`
- `DOCKERHUB_TOKEN`

## Step 4: Customize Deployment Settings (Optional)

Edit `.github/workflows/deploy-cloudrun.yml` to customize:

### Change Region
```yaml
env:
  REGION: us-central1  # Change to your preferred region
```

Available regions: `us-central1`, `us-east1`, `europe-west1`, `asia-northeast1`, etc.

### Change Service Name
```yaml
env:
  SERVICE_NAME: google-chat-bot  # Change to your preferred name
```

### Adjust Resources
In the deployment step, you can modify:
- `--memory 256Mi` - Increase if your app needs more memory (256Mi, 512Mi, 1Gi, 2Gi, 4Gi)
- `--cpu 1` - Number of CPUs (1, 2, 4, 8)
- `--min-instances 0` - Minimum instances (0 for scale-to-zero)
- `--max-instances 10` - Maximum instances
- `--timeout 300s` - Request timeout

### Authentication
The workflow sets `--allow-unauthenticated` which means anyone can access your service. To require authentication:

1. Remove or change the flag in `.github/workflows/deploy-cloudrun.yml`:
```yaml
--allow-unauthenticated  # Remove this line
```

2. To require authentication:
```yaml
--no-allow-unauthenticated
```

## Step 5: Deploy

Once everything is configured, deployment happens automatically:

1. **Automatic Deployment**: Push to the `main` branch
   ```bash
   git add .
   git commit -m "Add Cloud Run deployment"
   git push origin main
   ```

2. **Manual Deployment**: Go to GitHub Actions tab and trigger the "Deploy to Cloud Run" workflow manually

## Step 6: Verify Deployment

After deployment completes:

1. Check the GitHub Actions logs for the service URL
2. Or run locally:
   ```bash
   docker run --rm --volumes-from gcloud-config gcr.io/google.com/cloudsdktool/google-cloud-cli:stable \
   gcloud run services describe google-chat-bot \
     --platform managed \
     --region us-central1 \
     --project=YOUR_PROJECT_ID \
     --format 'value(status.url)'
   ```

3. Visit the URL to see your app

## Troubleshooting

### Permission Denied Errors
- Verify the service account has `roles/run.admin` and `roles/iam.serviceAccountUser`
- Check that Workload Identity Federation is configured correctly

### Image Pull Errors
- Ensure your Docker Hub image is public, or configure Cloud Run to authenticate with Docker Hub
- Verify the image name in the workflow matches your Docker Hub repository

### Environment Variable Issues
- Check that `GOOGLE_CHAT_WEBHOOK_URL` secret is set in GitHub
- Verify the webhook URL format is correct

### Workload Identity Federation Issues
- Ensure the `attribute.repository` condition matches your repository exactly
- Check that the repository owner matches what you configured
- Verify the service account has the `roles/iam.workloadIdentityUser` binding

### Docker gcloud Issues
- **"Container not found" error**: Make sure you ran the Step 0 authentication command to create the `gcloud-config` container
- **Authentication expired**: Re-run the authentication command from Step 0.1
- **Wrong project**: Check your project ID with:
  ```bash
  docker run --rm --volumes-from gcloud-config gcr.io/google.com/cloudsdktool/google-cloud-cli:stable \
  gcloud config get-value project
  ```

## Cleanup

When you're done with the setup and want to remove the gcloud Docker container:

```bash
# Remove the gcloud-config container
docker rm gcloud-config
```

Note: You can recreate it anytime by running the Step 0.1 authentication command again.

## Using Google Artifact Registry (Alternative)

If you prefer to use Google Artifact Registry instead of Docker Hub:

1. Enable Artifact Registry API:
   ```bash
   docker run --rm --volumes-from gcloud-config gcr.io/google.com/cloudsdktool/google-cloud-cli:stable \
   gcloud services enable artifactregistry.googleapis.com --project=YOUR_PROJECT_ID
   ```

2. Create a repository:
   ```bash
   docker run --rm --volumes-from gcloud-config gcr.io/google.com/cloudsdktool/google-cloud-cli:stable \
   gcloud artifacts repositories create google-chat-bot \
     --repository-format=docker \
     --location=us-central1 \
     --project=YOUR_PROJECT_ID
   ```

3. Update the workflow to push to Artifact Registry and deploy from there

## Cost Considerations

Cloud Run pricing (as of 2024):
- **CPU**: $0.00002400 per vCPU-second
- **Memory**: $0.00000250 per GiB-second
- **Requests**: $0.40 per million requests
- **Free tier**: 2 million requests per month, 360,000 GiB-seconds of memory, 180,000 vCPU-seconds

With `--min-instances 0`, your service scales to zero when not in use, minimizing costs.

## Additional Resources

- [Cloud Run Documentation](https://cloud.google.com/run/docs)
- [Workload Identity Federation Guide](https://cloud.google.com/iam/docs/workload-identity-federation)
- [GitHub Actions with GCP](https://github.com/google-github-actions)
