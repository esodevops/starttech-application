[![Frontend CI/CD](https://github.com/esodevops/starttech-application/actions/workflows/frontend-ci-cd.yml/badge.svg)](https://github.com/esodevops/starttech-application/actions/workflows/frontend-ci-cd.yml) [![Backend CI/CD](https://github.com/esodevops/starttech-application/actions/workflows/backend-ci-cd.yml/badge.svg)](https://github.com/esodevops/starttech-application/actions/workflows/backend-ci-cd.yml)

# StartTech Application

![StartTech Project](../starttech-img.png)

This is the full-stack application for StartTech, forked from [much-to-do](https://github.com/Innocent9712/much-to-do/tree/feature/full-stack).

It contains:

- **frontend/** — React app, deployed to S3 + CloudFront
- **backend/** — Golang API, deployed to EC2 via Docker and an Auto Scaling Group
- **scripts/** — Helper scripts for manual deploys and rollbacks

---

## Architecture

```
User → CloudFront CDN → S3 (React static files)
User → Application Load Balancer → EC2 ASG (Golang API) → MongoDB Atlas
                                                         → ElastiCache Redis
```

---

## CI/CD Pipelines

### Frontend (`frontend-ci-cd.yml`)

Triggered when files change in `frontend/`.

| Step            | What it does                          |
| --------------- | ------------------------------------- |
| Install & test  | `npm ci` + `npm test`                 |
| Security scan   | `npm audit --audit-level=high`        |
| Build           | `npm run build`                       |
| Deploy to S3    | `aws s3 sync build/ s3://your-bucket` |
| Clear CDN cache | CloudFront cache invalidation         |

### Backend (`backend-ci-cd.yml`)

Triggered when files change in `backend/`.

| Step               | What it does                                       |
| ------------------ | -------------------------------------------------- |
| Go tests           | `go test -race ./...`                              |
| Build Docker image | `docker build`                                     |
| Vulnerability scan | Trivy scans for HIGH/CRITICAL CVEs                 |
| Push to Docker Hub | Tagged with git SHA and `latest`                   |
| Rolling deploy     | ASG instance refresh (replaces EC2s one at a time) |
| Smoke test         | `curl /health` on the ALB                          |

---

## Required GitHub Secrets

Go to **Settings → Secrets and variables → Actions** and add:

| Secret                       | Description                |
| ---------------------------- | -------------------------- |
| `AWS_ACCESS_KEY_ID`          | IAM user access key        |
| `AWS_SECRET_ACCESS_KEY`      | IAM user secret key        |
| `DOCKERHUB_USERNAME`         | Docker Hub username        |
| `DOCKERHUB_TOKEN`            | Docker Hub access token    |
| `S3_BUCKET_NAME`             | Frontend S3 bucket name    |
| `CLOUDFRONT_DISTRIBUTION_ID` | CloudFront distribution ID |
| `REACT_APP_API_URL`          | Backend ALB URL            |

---

## Local Development

**Frontend:**

```bash
cd frontend
npm install
REACT_APP_API_URL=http://localhost:8080 npm start
```

**Backend:**

````bash
cd backend
# StartTech Application
This document provides a comprehensive guide for setting up, deploying, and managing the StartTech Application frontend and backend.

## Prerequisites
- Node.js (for frontend)
- Go (for backend)
- AWS CLI (for infrastructure management)
- Docker (optional, for containerized deployment)

## Directory Structure
- `frontend/` — React application
- `backend/` — Go backend service
- `scripts/` — Deployment and health check scripts

## Setup
### Frontend
1. Navigate to `frontend/`
2. Install dependencies:
  ```sh
  npm install
````

3. Start development server:

```sh
npm start
```

### Backend

1. Navigate to `backend/cmd/server/`
2. Build and run:

```sh
go build -o server main.go
./server
```

## Deployment

- Use the scripts in `scripts/` for deployment and health checks.
- Infrastructure is managed via the [`starttech-infra`](../starttech-infra/README.md) repository. Ensure both repos are cloned and up to date.

### Connecting Application and Infrastructure Repos

1. Deploy infrastructure using the `starttech-infra` repo (see its README for details).
2. After infra deploy, copy the output URLs (CloudFront, ALB) and update the relevant environment variables in this repo (e.g., `REACT_APP_API_URL`).
3. Use the provided secret sync scripts or GitHub Actions to propagate outputs (like S3 bucket, CloudFront ID, backend URL) between repos as secrets.

### Connecting MongoDB Atlas to GitHub Actions

To automate allowlisting of NAT Gateway IPs in MongoDB Atlas:

1. Create an Atlas API key with Project Owner permissions.
2. Add the following secrets to your GitHub repo:

- `ATLAS_PUBLIC_KEY`
- `ATLAS_PRIVATE_KEY`
- `ATLAS_PROJECT_ID`

3. The CI workflow will automatically update the Atlas IP Access List with the current NAT Gateway IP and remove old ones.

### Connecting Docker Image to Backend

1. The backend is containerized. The CI/CD pipeline builds and pushes the Docker image to Docker Hub.
2. The EC2 Auto Scaling Group pulls the latest image using the credentials provided in GitHub Secrets (`DOCKERHUB_USERNAME`, `DOCKERHUB_TOKEN`).
3. To run locally:

```sh
docker build -t much-to-do-backend ./backend
docker run -p 8080:8080 \
  -e MONGO_URI="<your-mongo-uri>" \
  -e REDIS_ADDR="localhost:6379" \
  much-to-do-backend
```

### IAM and Secret Propagation

- Use the least-privilege IAM policies from the infra repo for CI/CD users.
- Use the provided scripts or GitHub Actions to sync secrets and outputs between repos.

## Environment Variables

- Configure environment variables as needed for both frontend and backend.

## Additional Resources

- See `ARCHITECTURE.md` for system design.
- See `RUNBOOK.md` for operations and troubleshooting.
