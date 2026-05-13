#!/bin/bash
# -------------------------------------------------
# deploy-frontend.sh
#
# Builds the frontend app and uploads it to S3, then
# clears the CloudFront cache.
#
# Required environment variables:
#   S3_BUCKET_NAME              — your S3 bucket name
#   CLOUDFRONT_DISTRIBUTION_ID  — your CloudFront ID (optional for upload)
#   REACT_APP_API_URL           — optional API URL override (legacy, still supported)
#
# Usage:
#   export S3_BUCKET_NAME=starttech-frontend-prod
#   export CLOUDFRONT_DISTRIBUTION_ID=EXXXXXXXXXX
#   export REACT_APP_API_URL=http://starttech-alb-xxx.us-east-1.elb.amazonaws.com
#   ./scripts/deploy-frontend.sh
# -------------------------------------------------

set -e

# Validate required variables
: "${S3_BUCKET_NAME:?Set S3_BUCKET_NAME}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/../frontend"

echo "=== Building frontend app ==="
npm ci

# Optional explicit API URL override.
# If HTTPS frontend is used, injecting an HTTP API URL causes mixed-content failures.
if [ -n "${REACT_APP_API_URL:-}" ]; then
  if [[ "$REACT_APP_API_URL" == http://* ]]; then
    echo "Warning: REACT_APP_API_URL is HTTP. Ignoring it to avoid mixed-content errors."
    unset VITE_API_BASE_URL
  else
    export VITE_API_BASE_URL="$REACT_APP_API_URL"
  fi
fi

npm run build

echo ""
echo "=== Uploading to S3 ==="
# Upload everything except index.html with long cache (content-hashed filenames)
aws s3 sync dist/ "s3://$S3_BUCKET_NAME" \
  --delete \
  --exclude "index.html" \
  --cache-control "max-age=31536000,public"

# Upload index.html with no-cache so browsers always get the latest shell
aws s3 cp dist/index.html "s3://$S3_BUCKET_NAME/index.html" \
  --cache-control "no-cache,no-store,must-revalidate"

echo ""
echo "=== Invalidating CloudFront cache ==="
if [ -z "$CLOUDFRONT_DISTRIBUTION_ID" ]; then
  echo "Skipping invalidation: CLOUDFRONT_DISTRIBUTION_ID is not set."
  echo "Frontend files were uploaded to S3 successfully."
else
  set +e
  INVALIDATION_OUTPUT=$(aws cloudfront create-invalidation \
    --distribution-id "$CLOUDFRONT_DISTRIBUTION_ID" \
    --paths "/*" 2>&1)
  INVALIDATION_EXIT=$?
  set -e

  if [ "$INVALIDATION_EXIT" -ne 0 ]; then
    echo "$INVALIDATION_OUTPUT"
    if echo "$INVALIDATION_OUTPUT" | grep -q "NoSuchDistribution"; then
      echo ""
      echo "CloudFront distribution ID is stale or from another account."
      echo "Update CLOUDFRONT_DISTRIBUTION_ID from starttech-infra Terraform output and retry."
    elif echo "$INVALIDATION_OUTPUT" | grep -q "AccessDenied"; then
      echo ""
      echo "IAM user lacks CloudFront permissions."
      echo "Grant cloudfront:CreateInvalidation and cloudfront:GetDistribution on the distribution ARN."
    fi
    echo "Continuing without invalidation."
  else
    echo "$INVALIDATION_OUTPUT"
  fi
fi

echo ""
if [ -n "$CLOUDFRONT_DISTRIBUTION_ID" ]; then
  set +e
  DISTRIBUTION_DOMAIN=$(aws cloudfront get-distribution \
    --id "$CLOUDFRONT_DISTRIBUTION_ID" \
    --query "Distribution.DomainName" \
    --output text 2>/dev/null)
  LOOKUP_EXIT=$?
  set -e

  if [ "$LOOKUP_EXIT" -eq 0 ] && [ -n "$DISTRIBUTION_DOMAIN" ] && [ "$DISTRIBUTION_DOMAIN" != "None" ]; then
    echo "Frontend deployed! Visit: https://$DISTRIBUTION_DOMAIN"
  else
    echo "Frontend deployed to S3 bucket: s3://$S3_BUCKET_NAME"
  fi
else
  echo "Frontend deployed to S3 bucket: s3://$S3_BUCKET_NAME"
fi
