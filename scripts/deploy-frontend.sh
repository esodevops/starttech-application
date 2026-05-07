#!/bin/bash
# -------------------------------------------------
# deploy-frontend.sh
#
# Builds the React app and uploads it to S3, then
# clears the CloudFront cache.
#
# Required environment variables:
#   S3_BUCKET_NAME              — your S3 bucket name
#   CLOUDFRONT_DISTRIBUTION_ID  — your CloudFront ID
#   REACT_APP_API_URL           — backend ALB URL
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
: "${CLOUDFRONT_DISTRIBUTION_ID:?Set CLOUDFRONT_DISTRIBUTION_ID}"
: "${REACT_APP_API_URL:?Set REACT_APP_API_URL}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/../frontend"

echo "=== Building React app ==="
npm ci
npm run build

echo ""
echo "=== Uploading to S3 ==="
# Upload everything except index.html with long cache (content-hashed filenames)
aws s3 sync build/ "s3://$S3_BUCKET_NAME" \
  --delete \
  --exclude "index.html" \
  --cache-control "max-age=31536000,public"

# Upload index.html with no-cache so browsers always get the latest shell
aws s3 cp build/index.html "s3://$S3_BUCKET_NAME/index.html" \
  --cache-control "no-cache,no-store,must-revalidate"

echo ""
echo "=== Invalidating CloudFront cache ==="
aws cloudfront create-invalidation \
  --distribution-id "$CLOUDFRONT_DISTRIBUTION_ID" \
  --paths "/*"

echo ""
echo "Frontend deployed! Visit: https://$(aws cloudfront get-distribution \
  --id "$CLOUDFRONT_DISTRIBUTION_ID" \
  --query "Distribution.DomainName" \
  --output text)"
