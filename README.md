# StartTech Application

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

```bash
cd backend
export MONGO_URI="mongodb+srv://..."
export REDIS_ADDR="localhost:6379"
go run ./cmd/server
```

**Run with Docker:**

```bash
docker build -t much-to-do-backend ./backend
docker run -p 8080:8080 \
  -e MONGO_URI="$MONGO_URI" \
  -e REDIS_ADDR="localhost:6379" \
  much-to-do-backend
```

---

## Manual Deploy Scripts

```bash
# Deploy frontend
export S3_BUCKET_NAME=starttech-frontend-prod
export CLOUDFRONT_DISTRIBUTION_ID=EXXXXXXXXXX
export REACT_APP_API_URL=http://your-alb.us-east-1.elb.amazonaws.com
./scripts/deploy-frontend.sh

# Deploy backend
export DOCKERHUB_USERNAME=myuser
export DOCKERHUB_TOKEN=dckr_pat_xxx
export IMAGE_TAG=v1.0.0
./scripts/deploy-backend.sh

# Rollback backend to a previous version
./scripts/rollback.sh v0.9.0

# Health check
./scripts/health-check.sh http://your-alb.us-east-1.elb.amazonaws.com
```
