# StartTech Application — Operations & Troubleshooting Runbook

Operational guide for developers and on-call engineers managing the StartTech application in development and production.

---

## Service inventory

| Service | URL / identifier | Owner repo |
|---------|------------------|------------|
| Frontend (prod) | `https://<cloudfront-domain>` | starttech-application |
| Backend ALB | `http://<alb-dns>` | starttech-infra |
| Health endpoint | `<backend>/health` | starttech-application |
| Swagger | `<backend>/swagger/index.html` | starttech-application |
| CloudWatch logs | `/starttech/prod/backend` | starttech-infra |
| Docker image | `<dockerhub-user>/much-to-do-backend` | starttech-application |
| ASG | `starttech-asg` (default) | starttech-infra |

---

## Daily operations

### Check system health

```bash
# Backend
curl -s http://<ALB_DNS>/health | jq .

# Expected (healthy):
# { "database": "ok", "cache": "ok" }     # cache ok when Redis enabled
# { "database": "ok", "cache": "disabled" } # local / cache off
```

```bash
# Frontend (via CloudFront)
curl -sI https://<CLOUDFRONT_DOMAIN>/ | head -5
```

### View recent backend logs

AWS Console → **CloudWatch → Log groups → `/starttech/prod/backend`**

Or CLI:

```bash
aws logs tail /starttech/prod/backend --follow --region us-east-1
```

Log stream name: `backend-logs` (Docker awslogs driver).

Use queries from `starttech-infra/monitoring/log-insights-queries.txt`.

---

## Deployment runbooks

### Deploy frontend (CI or manual)

**Automated:** Push to `main` with changes under `frontend/`.

**Manual:**

```bash
export S3_BUCKET_NAME="<from-terraform-output>"
export CLOUDFRONT_DISTRIBUTION_ID="<from-terraform-output>"
./scripts/deploy-frontend.sh
```

**Verify:**

1. Open CloudFront URL in browser (hard refresh or incognito).
2. Register/login and load `/todos`.
3. Check browser Network tab — API calls should hit same origin (CloudFront), not raw ALB HTTP.

### Deploy backend (CI or manual)

**Automated:** Push to `main` with changes under `backend/`.

Pipeline: test → Docker build → push to Docker Hub → update launch template image tag → ASG instance refresh → smoke test on `/health`.

**Manual:**

```bash
export DOCKERHUB_USERNAME="<user>"
export DOCKERHUB_TOKEN="<token>"
export IMAGE_TAG="$(git rev-parse HEAD)"
./scripts/deploy-backend.sh
```

**Verify:**

```bash
./scripts/health-check.sh http://<ALB_DNS>
aws autoscaling describe-instance-refreshes \
  --auto-scaling-group-name starttech-asg \
  --query 'InstanceRefreshes[0].[Status,PercentageComplete]' \
  --output table
```

### Rollback backend

```bash
# Option 1: Re-run pipeline on previous commit (recommended)

# Option 2: Manual — push previous image tag to launch template and refresh ASG
# Update user-data image reference to known-good SHA, then:
aws autoscaling start-instance-refresh \
  --auto-scaling-group-name starttech-asg \
  --strategy Rolling \
  --preferences '{"MinHealthyPercentage": 50, "InstanceWarmup": 120}'
```

See `scripts/rollback.sh` for helper steps.

---

## Verify Redis caching

Caching is **enabled in production** via EC2 user-data (`ENABLE_CACHE=true`).

### Step 1 — Health endpoint

```bash
curl -s http://<ALB_DNS>/health
```

| `cache` value | Meaning |
|---------------|---------|
| `ok` | Redis reachable |
| `down` | Enabled but ping failed |
| `disabled` | `ENABLE_CACHE` not true (misconfigured instance) |

### Step 2 — Log evidence

After API traffic, search CloudWatch for:

```
[CACHE HIT]
[CACHE MISS]
[CACHE SET]
```

**Username check test:**

```bash
curl -s "http://<ALB_DNS>/auth/username-check/testuser123"
# Repeat immediately — second request should log [CACHE HIT] if user was cached as taken
```

**Todos test (requires auth):**

1. `GET /tasks` twice — expect MISS then HIT in logs.
2. `POST /tasks` — expect `[CACHE DELETE]` for user's list key.

### Step 3 — ElastiCache metrics

CloudWatch → **ElastiCache** → cluster `starttech-redis`:

- `CurrConnections` > 0
- `CacheHitRate` increases under repeated reads

### Common cache failures

| Symptom | Likely cause | Fix |
|---------|--------------|-----|
| `cache: disabled` | Old EC2 without `ENABLE_CACHE` | Terraform apply + ASG instance refresh |
| `cache: down` | SG blocks 6379, wrong endpoint | Check Redis SG and `REDIS_ADDR` in user-data |
| Always MISS | NoOpCache fallback | Check startup logs for Redis connection error |
| App crash on Redis error | Old image with `log.Fatalf` | Deploy latest backend code |

---

## Verify CloudWatch logging

### Expected configuration

- **Log group:** `/starttech/prod/backend`
- **Log stream:** `backend-logs`
- **Source:** Docker container stdout/stderr via `awslogs` driver

### Verification

```bash
aws logs describe-log-streams \
  --log-group-name "/starttech/prod/backend" \
  --order-by LastEventTime \
  --descending \
  --limit 5

aws logs tail /starttech/prod/backend --since 10m
```

Look for startup messages:

- `Logger initialized`
- `Successfully connected to MongoDB`
- `Caching is enabled. Connecting to Redis...`
- `Successfully connected to Redis.`

### No logs appearing

| Check | Action |
|-------|--------|
| Container running? | SSM into instance: `docker ps`, `docker logs backend` |
| IAM permissions | EC2 role needs CloudWatch Logs write (managed policy attached) |
| Log group exists? | Terraform monitoring module creates it |
| Wrong group name? | User-data `cloudwatch_log_group` must match Terraform |
| ASG not refreshed | Old instances won't have awslogs config |

---

## Troubleshooting guide

### Frontend: blank page or 403 on refresh

**Cause:** SPA routing — S3/CloudFront must return `index.html` for unknown paths.

**Fix:** CloudFront custom error responses (403/404 → 200 `/index.html`) are in infra Terraform. Re-apply storage module if missing.

### Frontend: API calls fail (CORS / mixed content)

**Symptoms:** Browser blocks `http://` ALB from `https://` CloudFront page.

**Fix:**

- Leave `VITE_API_BASE_URL` empty in production build (same-origin proxy).
- Ensure `alb_dns_name` is set in infra so CloudFront proxies `/auth/*`, `/tasks/*`, `/health`.

### Backend: 502/503 from ALB

**Checks:**

1. `curl http://<ALB>/health` — timeout or 503?
2. Target group health in EC2 console — instances healthy?
3. Security groups — ALB → backend on 8080?
4. `docker logs backend` on instance — MongoDB connection errors?

**MongoDB Atlas:**

- Confirm NAT Gateway public IP is in Atlas IP Access List (infra workflow automates this).
- Test from instance: outbound connectivity to Atlas host.

### Backend: 401 on protected routes

- Cookie `Secure` flag mismatch (HTTP vs HTTPS).
- `COOKIE_DOMAINS` / `ALLOWED_ORIGINS` misconfigured.
- Expired JWT — re-login.
- Frontend not sending credentials — `withCredentials: true` required.

### Backend: high latency

1. CloudWatch ALB `TargetResponseTime` alarm (infra).
2. Check MongoDB Atlas metrics (slow queries).
3. Redis cache hit rate — low hits mean more DB load.
4. ASG CPU — scale-out policy may need tuning.

### CI/CD: backend deploy fails

| Error | Resolution |
|-------|------------|
| Docker Hub auth | Verify `DOCKERHUB_USERNAME` / `DOCKERHUB_TOKEN` secrets |
| ASG not found | Set `ASG_NAME` secret or ensure `starttech-asg` exists |
| Instance refresh stuck | Check unhealthy instances, launch template, user-data logs |
| Smoke test fails | ALB `/health` not 200 — debug running containers |

### CI/CD: frontend deploy fails

| Error | Resolution |
|-------|------------|
| S3 access denied | IAM policy for bucket sync |
| Wrong bucket | Update `S3_BUCKET_NAME` from Terraform output |
| Stale UI after deploy | CloudFront invalidation — check `CLOUDFRONT_DISTRIBUTION_ID` |

---

## Integration tests (local)

```bash
cd backend
export INTEGRATION=true
go test -tags=integration -v ./internal/handlers/...
```

Requires Docker (testcontainers for MongoDB + Redis).

---

## Escalation checklist

When production is impaired, collect:

1. `/health` response from ALB
2. Last 50 lines from CloudWatch log group `/starttech/prod/backend`
3. ALB target health state
4. Recent GitHub Actions workflow run (frontend + backend + infra)
5. Atlas cluster status and IP allowlist
6. ElastiCache cluster status

---

## Related documentation

- [README.md](./README.md) — Setup and deployment
- [ARCHITECTURE.md](./ARCHITECTURE.md) — System design
- [starttech-infra RUNBOOK.md](../starttech-infra/RUNBOOK.md) — Infrastructure operations
