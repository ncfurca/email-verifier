# Email Verifier API Deployment Guide

This guide provides detailed instructions for deploying the Email Verifier API to various cloud platforms.

## Table of Contents

1. [Preparing for Deployment](#preparing-for-deployment)
2. [Deploying to Render](#deploying-to-render)
3. [Deploying to Google Cloud Run](#deploying-to-google-cloud-run)
4. [Deploying to AWS Elastic Beanstalk](#deploying-to-aws-elastic-beanstalk)
5. [Deploying to DigitalOcean App Platform](#deploying-to-digitalocean-app-platform)
6. [Post-Deployment Configuration](#post-deployment-configuration)
7. [Monitoring and Maintenance](#monitoring-and-maintenance)

## Preparing for Deployment

Before deploying the Email Verifier API, you need to prepare your codebase and environment.

### Prerequisites

- Git repository with your Email Verifier code
- Go 1.18 or higher installed locally
- Basic familiarity with command line tools
- Account on your chosen cloud platform (Render, Google Cloud, AWS, or DigitalOcean)

### Testing Locally

Before deploying, test your API locally to ensure it works correctly:

1. Build the API server:
   ```bash
   go build -o api ./cmd/api
   ```

2. Run the API server:
   ```bash
   ./api -port 8080
   ```

3. Test the API with curl:
   ```bash
   curl -X GET http://localhost:8080/health
   ```

   You should receive a response like:
   ```json
   {"status":"ok","time":"2023-05-15T12:34:56Z"}
   ```

### Preparing Configuration

Create a production-ready `config.yaml` file:

```yaml
# Input/Output Configuration
input_file: "input.csv"
input_type: "csv"
output_file: "output.csv"
output_type: "csv"

# Verification Settings
valid_threshold: 80
risky_threshold: 60
default_risky_score: 50
max_retries: 3
initial_backoff: 1s
num_workers: 10

# Scoring Weights
scoring_weights:
  has_mx_records: 20
  reachable_yes: 40
  reachable_unknown: 20
  role_account: -10
  free_provider: -5
  suggestion: -10
```

## Deploying to Render

[Render](https://render.com/) is a unified cloud platform that makes it easy to deploy web services with zero DevOps.

### Step 1: Create a Render Account

If you don't already have one, sign up for a Render account at [render.com](https://render.com/).

### Step 2: Connect Your Repository

1. From the Render dashboard, click **New** and select **Web Service**
2. Connect your GitHub/GitLab account or use a public repository URL
3. Select your Email Verifier repository

### Step 3: Configure the Web Service

Configure your web service with the following settings:

- **Name**: `email-verifier-api` (or your preferred name)
- **Environment**: `Go`
- **Region**: Choose the region closest to your users
- **Branch**: `main` (or your default branch)
- **Build Command**: `go build -o api ./cmd/api`
- **Start Command**: `./api -port $PORT`
- **Plan**: Choose the appropriate plan (Free tier works for testing)

### Step 4: Add Environment Variables

Add the following environment variables:

- `PORT`: `10000` (Render will automatically set this, but it's good to specify)

### Step 5: Deploy

Click **Create Web Service** to deploy your API. Render will automatically build and deploy your application.

### Step 6: Access Your API

Once deployed, your API will be available at:
```
https://email-verifier-api.onrender.com
```

Test it with:
```bash
curl -X GET https://email-verifier-api.onrender.com/health
```

## Deploying to Google Cloud Run

[Google Cloud Run](https://cloud.google.com/run) is a fully managed platform for containerized applications.

### Step 1: Set Up Google Cloud

1. Create a [Google Cloud account](https://cloud.google.com/) if you don't have one
2. Install the [Google Cloud SDK](https://cloud.google.com/sdk/docs/install)
3. Initialize the SDK:
   ```bash
   gcloud init
   ```

### Step 2: Create a Dockerfile

Create a `Dockerfile` in your project root:

```dockerfile
FROM golang:1.18-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o api ./cmd/api

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/api /app/
COPY config.yaml /app/
EXPOSE 8080
CMD ["/app/api", "-port", "8080"]
```

### Step 3: Build and Push the Docker Image

1. Enable the Container Registry API:
   ```bash
   gcloud services enable containerregistry.googleapis.com
   ```

2. Build and push the Docker image:
   ```bash
   gcloud builds submit --tag gcr.io/YOUR_PROJECT_ID/email-verifier
   ```
   Replace `YOUR_PROJECT_ID` with your Google Cloud project ID.

### Step 4: Deploy to Cloud Run

Deploy the image to Cloud Run:

```bash
gcloud run deploy email-verifier \
  --image gcr.io/YOUR_PROJECT_ID/email-verifier \
  --platform managed \
  --allow-unauthenticated \
  --region us-central1
```

Replace `YOUR_PROJECT_ID` with your Google Cloud project ID and choose an appropriate region.

### Step 5: Access Your API

Once deployed, your API will be available at the URL provided in the deployment output. Test it with:

```bash
curl -X GET https://email-verifier-HASH.run.app/health
```

Replace `HASH` with the unique identifier in your Cloud Run URL.

## Deploying to AWS Elastic Beanstalk

[AWS Elastic Beanstalk](https://aws.amazon.com/elasticbeanstalk/) is an easy-to-use service for deploying and scaling web applications.

### Step 1: Set Up AWS

1. Create an [AWS account](https://aws.amazon.com/) if you don't have one
2. Install the [AWS CLI](https://aws.amazon.com/cli/) and [EB CLI](https://docs.aws.amazon.com/elasticbeanstalk/latest/dg/eb-cli3-install.html)
3. Configure the AWS CLI:
   ```bash
   aws configure
   ```

### Step 2: Initialize Elastic Beanstalk

1. Navigate to your project directory
2. Initialize Elastic Beanstalk:
   ```bash
   eb init
   ```
3. Follow the prompts to configure your application:
   - Select a region
   - Create a new application or use an existing one
   - Select "Go" as the platform
   - Choose whether to use CodeCommit (optional)
   - Set up SSH for instance access (optional)

### Step 3: Create Configuration Files

1. Create a `Procfile` in your project root:
   ```
   web: ./api -port $PORT
   ```

2. Create a `.ebextensions/01_build.config` file:
   ```yaml
   commands:
     01_build:
       command: "go build -o api ./cmd/api"
   ```

3. Create a `.ebextensions/02_files.config` file:
   ```yaml
   files:
     "/opt/elasticbeanstalk/hooks/appdeploy/post/99_copy_config.sh":
       mode: "000755"
       owner: root
       group: root
       content: |
         #!/bin/bash
         cp /var/app/current/config.yaml /var/app/current/
   ```

### Step 4: Deploy Your Application

Create and deploy your Elastic Beanstalk environment:

```bash
eb create email-verifier-env
```

### Step 5: Access Your API

Once deployed, your API will be available at the URL provided in the deployment output. Test it with:

```bash
curl -X GET http://email-verifier-env.REGION.elasticbeanstalk.com/health
```

Replace `REGION` with your AWS region identifier.

## Deploying to DigitalOcean App Platform

[DigitalOcean App Platform](https://www.digitalocean.com/products/app-platform/) is a Platform-as-a-Service (PaaS) offering that allows you to build, deploy, and scale apps quickly.

### Step 1: Set Up DigitalOcean

1. Create a [DigitalOcean account](https://www.digitalocean.com/) if you don't have one
2. Install the [doctl CLI](https://docs.digitalocean.com/reference/doctl/how-to/install/) (optional)

### Step 2: Create a New App

1. From the DigitalOcean dashboard, click **Create** and select **Apps**
2. Connect your GitHub/GitLab account and select your repository
3. Select the branch you want to deploy

### Step 3: Configure Your App

Configure your app with the following settings:

- **Type**: Web Service
- **Source Directory**: `/`
- **Build Command**: `go build -o api ./cmd/api`
- **Run Command**: `./api -port $PORT`
- **HTTP Port**: `8080`

### Step 4: Add Environment Variables

Add the following environment variables:

- `PORT`: `8080`

### Step 5: Deploy Your App

Click **Next** and then **Create Resources** to deploy your app.

### Step 6: Access Your API

Once deployed, your API will be available at the URL provided in the deployment output. Test it with:

```bash
curl -X GET https://email-verifier-HASH.ondigitalocean.app/health
```

Replace `HASH` with the unique identifier in your DigitalOcean App URL.

## Post-Deployment Configuration

After deploying your API, you may need to perform additional configuration.

### Setting Up a Custom Domain

For a more professional appearance, set up a custom domain for your API:

1. Purchase a domain from a domain registrar (e.g., Namecheap, GoDaddy)
2. Follow your cloud provider's instructions to add a custom domain:
   - [Render Custom Domains](https://render.com/docs/custom-domains)
   - [Google Cloud Run Custom Domains](https://cloud.google.com/run/docs/mapping-custom-domains)
   - [AWS Elastic Beanstalk Custom Domains](https://docs.aws.amazon.com/elasticbeanstalk/latest/dg/customdomains.html)
   - [DigitalOcean App Platform Custom Domains](https://docs.digitalocean.com/products/app-platform/how-to/manage-domains/)

### Setting Up HTTPS

Most cloud providers automatically configure HTTPS for your API. If not, follow your provider's instructions to set up HTTPS with Let's Encrypt or another SSL certificate provider.

### Configuring CORS

The API includes built-in CORS support. If you need to restrict access to specific domains, modify the `corsMiddleware` function in `pkg/api/server.go`:

```go
func corsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Replace "*" with your allowed domains
        w.Header().Set("Access-Control-Allow-Origin", "https://yourdomain.com")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

## Monitoring and Maintenance

### Monitoring Your API

Set up monitoring to ensure your API is running smoothly:

1. Use your cloud provider's built-in monitoring tools:
   - [Render Metrics](https://render.com/docs/metrics)
   - [Google Cloud Monitoring](https://cloud.google.com/monitoring)
   - [AWS CloudWatch](https://aws.amazon.com/cloudwatch/)
   - [DigitalOcean Monitoring](https://docs.digitalocean.com/products/monitoring/)

2. Set up uptime monitoring with a service like:
   - [UptimeRobot](https://uptimerobot.com/) (free tier available)
   - [Pingdom](https://www.pingdom.com/)
   - [StatusCake](https://www.statuscake.com/)

### Updating Your API

When you need to update your API:

1. Make and test changes locally
2. Commit changes to your repository
3. Deploy using your cloud provider's deployment process:
   - Render: Automatic deployment on push to the configured branch
   - Google Cloud Run: Run the build and deploy commands again
   - AWS Elastic Beanstalk: Run `eb deploy`
   - DigitalOcean: Automatic deployment on push to the configured branch

### Scaling Your API

If your API usage grows, you may need to scale:

1. Vertical scaling (more powerful instances):
   - Upgrade to a higher tier on your cloud provider

2. Horizontal scaling (more instances):
   - Enable auto-scaling if available on your cloud provider
   - Increase the number of workers in your configuration

### Backup and Disaster Recovery

Implement backup and disaster recovery procedures:

1. Regularly back up your configuration files
2. Document your deployment process
3. Set up automated deployments with CI/CD
4. Consider multi-region deployments for high availability 