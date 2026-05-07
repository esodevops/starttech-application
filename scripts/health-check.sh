#!/bin/bash
# -------------------------------------------------
# health-check.sh
#
# Verifies the backend API is healthy by calling
# the /health endpoint on the ALB.
#
# Usage:
#   ./scripts/health-check.sh http://your-alb-dns.us-east-1.elb.amazonaws.com
# -------------------------------------------------

set -e

BASE_URL="${1:-http://localhost:8080}"

echo "Checking health at $BASE_URL/health ..."

HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/health")

if [ "$HTTP_STATUS" = "200" ]; then
  echo "Health check PASSED (HTTP $HTTP_STATUS)"
  exit 0
else
  echo "Health check FAILED (HTTP $HTTP_STATUS)"
  exit 1
fi
