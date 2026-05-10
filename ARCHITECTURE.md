# System Architecture

## Overview

The StartTech Application is composed of a React frontend and a Go backend, deployed on AWS infrastructure managed via Terraform.

## Components

- **Frontend:** React SPA served via S3/CloudFront
- **Backend:** Go API server running on EC2
- **Infrastructure:** AWS (VPC, EC2, S3, CloudFront, NAT Gateway, MongoDB Atlas, etc.)

## Data Flow

1. User interacts with the React frontend.
2. Frontend communicates with the Go backend via REST APIs.
3. Backend interacts with MongoDB Atlas and other AWS services.

## Security

- IAM least-privilege policies for CI/CD
- NAT Gateway for secure outbound traffic
- Secrets managed via GitHub Actions and AWS Secrets Manager

## CI/CD and Integration

- GitHub Actions for build, test, deploy, and Atlas allowlist automation
- Cross-repo secret propagation: Infrastructure outputs (e.g., backend URL, S3 bucket) are synced from the infra repo to this repo as GitHub secrets, enabling seamless deployment and configuration.
- MongoDB Atlas allowlist automation: The CI workflow uses Atlas API keys (stored as secrets) to update the IP Access List with the current NAT Gateway IP and remove old ones automatically.
- Docker image flow: Backend Docker images are built and pushed to Docker Hub by CI/CD, then pulled by EC2 instances during deployment using credentials from secrets.


