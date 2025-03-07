# Email Verifier API Documentation

This document provides detailed instructions for setting up, using, and deploying the Email Verifier API service.

## Table of Contents

1. [Overview](#overview)
2. [Getting Started](#getting-started)
3. [API Endpoints](#api-endpoints)
4. [Google Sheets Integration](#google-sheets-integration)
5. [Deployment Guide](#deployment-guide)
6. [Troubleshooting](#troubleshooting)
7. [Performance Considerations](#performance-considerations)

## Overview

The Email Verifier API is a RESTful service that provides email verification capabilities through HTTP endpoints. It leverages the same robust verification engine as the command-line tool but exposes the functionality as a web service that can be integrated with other applications, particularly Google Sheets.

Key features:
- Verify individual emails
- Batch verify multiple emails
- Special endpoint for Google Sheets integration
- CORS support for cross-origin requests
- Health check endpoint for monitoring

## Getting Started

### Prerequisites

- Go 1.18 or higher
- Internet connection for DNS and SMTP checks
- Port 25 open for SMTP verification (or use a proxy)
- Port 8080 (or your chosen port) available for the API server

### Building the API Server

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/email-verifier.git
   cd email-verifier
   ```

2. Install dependencies:
   ```bash
   go mod download
   go get github.com/gorilla/mux
   ```

3. Build the API server:
   ```bash
   go build -o api ./cmd/api
   ```

### Running the API Server

Run the API server with:

```bash
./api -port 8080 -config config.yaml
```

Command-line options:
- `-port`: The port to run the API server on (default: 8080)
- `-config`: Path to the configuration file (default: config.yaml)

If the configuration file is not found, the server will use default settings.

### Configuration

The API server uses the same configuration file as the command-line tool. Here's a sample configuration:

```yaml
# Input/Output Configuration
input_file: "leads.csv"
input_type: "csv"
output_file: "verified_leads.csv"
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

## API Endpoints

### Health Check

**Endpoint**: `GET /health`

Checks if the API server is running properly.

**Example Request**:
```bash
curl -X GET http://localhost:8080/health
```

**Example Response**:
```json
{
  "status": "ok",
  "time": "2023-05-15T12:34:56Z"
}
```

### Verify Single Email

**Endpoint**: `POST /verify`

Verifies a single email address.

**Request Body**:
```json
{
  "email": "example@example.com"
}
```

**Example Request**:
```bash
curl -X POST http://localhost:8080/verify \
  -H "Content-Type: application/json" \
  -d '{"email": "example@example.com"}'
```

**Example Response**:
```json
{
  "email": "example@example.com",
  "verification_status": "valid",
  "confidence_score": 85,
  "processed_at": "2023-05-15T12:34:56Z"
}
```

### Batch Verify Emails

**Endpoint**: `POST /batch-verify`

Verifies multiple email addresses in a single request.

**Request Body**:
```json
{
  "emails": [
    "example1@example.com",
    "example2@example.com",
    "example3@example.com"
  ]
}
```

**Example Request**:
```bash
curl -X POST http://localhost:8080/batch-verify \
  -H "Content-Type: application/json" \
  -d '{"emails": ["example1@example.com", "example2@example.com", "example3@example.com"]}'
```

**Example Response**:
```json
{
  "results": [
    {
      "email": "example1@example.com",
      "verification_status": "valid",
      "confidence_score": 85,
      "processed_at": "2023-05-15T12:34:56Z"
    },
    {
      "email": "example2@example.com",
      "verification_status": "risky",
      "confidence_score": 65,
      "processed_at": "2023-05-15T12:34:57Z"
    },
    {
      "email": "example3@example.com",
      "verification_status": "invalid",
      "confidence_score": 20,
      "processed_at": "2023-05-15T12:34:58Z"
    }
  ]
}
```

### Google Sheets Integration

**Endpoint**: `POST /google-sheets`

Special endpoint optimized for Google Sheets integration.

**Request Body**:
```json
{
  "emails": [
    "example1@example.com",
    "example2@example.com"
  ]
}
```

**Example Request**:
```bash
curl -X POST http://localhost:8080/google-sheets \
  -H "Content-Type: application/json" \
  -d '{"emails": ["example1@example.com", "example2@example.com"]}'
```

**Example Response**:
```json
{
  "results": [
    {
      "email": "example1@example.com",
      "verification_status": "valid",
      "confidence_score": 85,
      "processed_at": "2023-05-15T12:34:56Z"
    },
    {
      "email": "example2@example.com",
      "verification_status": "risky",
      "confidence_score": 65,
      "processed_at": "2023-05-15T12:34:57Z"
    }
  ]
}
```

## Google Sheets Integration

### Setup Instructions

1. Open your Google Sheet
2. Go to Extensions > Apps Script
3. Create a new script file and paste the contents of `pkg/api/google_sheets_client.js`
4. Update the `API_URL` constant at the top of the file to point to your deployed API:
   ```javascript
   const API_URL = "https://your-api-url.com/google-sheets";
   ```
5. Save the script and reload your Google Sheet
6. You'll see a new "Email Verifier" menu item in the menu bar

### Using the Google Sheets Add-on

#### Verify Selected Emails

1. Select a range of cells containing email addresses
2. Click on "Email Verifier" > "Verify Selected Emails"
3. Confirm the operation when prompted
4. The verification results will appear in adjacent columns

#### Verify an Entire Column

1. Click on "Email Verifier" > "Verify Column"
2. Enter the column letter containing the emails (e.g., "A")
3. Confirm the operation when prompted
4. The verification results will appear in adjacent columns

#### Configure API Settings

1. Click on "Email Verifier" > "Settings"
2. Enter your API URL or leave blank to use the default
3. Click "OK" to save the settings

### Example Workflow

1. Create a Google Sheet with a column of email addresses
2. Deploy the Email Verifier API to a cloud provider
3. Configure the Google Sheets Add-on to use your API
4. Use the "Verify Column" feature to verify all emails
5. The sheet will be updated with verification statuses and confidence scores
6. Filter or sort the results as needed

## Deployment Guide

### Deploying to Render

[Render](https://render.com/) is a cloud platform that makes it easy to deploy web services.

1. Create a new Web Service on Render
2. Connect your GitHub repository
3. Configure the build settings:
   - Build Command: `go build -o api ./cmd/api`
   - Start Command: `./api -port $PORT`
4. Add environment variables:
   - `PORT`: `10000` (or any port Render assigns)
5. Deploy the service

### Deploying to Google Cloud Run

[Google Cloud Run](https://cloud.google.com/run) is a serverless platform for containerized applications.

1. Install the [Google Cloud SDK](https://cloud.google.com/sdk/docs/install)
2. Create a Dockerfile in your project root:
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
3. Build and push the Docker image:
   ```bash
   gcloud builds submit --tag gcr.io/YOUR_PROJECT_ID/email-verifier
   ```
4. Deploy to Cloud Run:
   ```bash
   gcloud run deploy email-verifier \
     --image gcr.io/YOUR_PROJECT_ID/email-verifier \
     --platform managed \
     --allow-unauthenticated
   ```

### Deploying to AWS Elastic Beanstalk

1. Install the [AWS CLI](https://aws.amazon.com/cli/) and [EB CLI](https://docs.aws.amazon.com/elasticbeanstalk/latest/dg/eb-cli3-install.html)
2. Initialize your EB application:
   ```bash
   eb init
   ```
3. Create a `Procfile` in your project root:
   ```
   web: ./api -port $PORT
   ```
4. Create a `.ebextensions/01_build.config` file:
   ```yaml
   commands:
     01_build:
       command: "go build -o api ./cmd/api"
   ```
5. Deploy your application:
   ```bash
   eb create email-verifier-env
   ```

## Troubleshooting

### Common Issues

#### API Server Won't Start

- Check if the port is already in use
- Ensure you have the correct permissions to bind to the port
- Verify that the configuration file exists and is valid

#### SMTP Verification Timeouts

- Most ISPs block outgoing SMTP requests through port 25
- Consider using a SOCKS proxy for SMTP verification
- Adjust the timeout settings in the configuration

#### CORS Issues

- The API includes CORS headers for cross-origin requests
- If you're still experiencing CORS issues, check your client-side code
- Ensure the request includes the appropriate headers

#### Google Sheets Integration Not Working

- Verify that the API URL is correct in the Apps Script
- Check the browser console for any JavaScript errors
- Ensure the API server is accessible from Google's servers

### Logging

The API server logs all requests and errors to the console. To save logs to a file, redirect the output:

```bash
./api -port 8080 > api.log 2>&1
```

## Performance Considerations

### Rate Limiting

The API includes built-in rate limiting to prevent overwhelming SMTP servers. You can adjust the rate limit in the configuration file:

```yaml
rate_limit: 5
rate_limit_period: 1
```

This limits the API to 5 SMTP connections per second.

### Connection Pooling

The API uses connection pooling for SMTP connections to improve performance. You can adjust the maximum number of connections in the configuration:

```yaml
max_connections: 10
```

### Scaling

For high-volume deployments, consider:

1. Increasing the number of workers
2. Deploying multiple instances behind a load balancer
3. Using a caching layer for frequently verified domains
4. Implementing a queue system for batch processing

### Monitoring

Monitor your API server's performance using:

- The `/health` endpoint for basic health checks
- Prometheus metrics (if implemented)
- Cloud provider monitoring tools
- Log analysis for error patterns 