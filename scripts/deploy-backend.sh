#!/bin/bash
# -------------------------------------------------
# deploy-backend.sh
#
# Builds, pushes, and deploys the backend Docker
# image to EC2 via an ASG rolling instance refresh.
#
# Required environment variables:
#   DOCKERHUB_USERNAME  — Docker Hub username
#   DOCKERHUB_TOKEN     — Docker Hub access token
#   IMAGE_TAG           — tag to apply (e.g. v1.2.3 or git SHA)
#
# Usage:
#   export DOCKERHUB_USERNAME=myuser
#   export DOCKERHUB_TOKEN=dckr_pat_xxx
#   export IMAGE_TAG=v1.0.0
#   ./scripts/deploy-backend.sh
# -------------------------------------------------

set -e

: "${DOCKERHUB_USERNAME:?Set DOCKERHUB_USERNAME}"
: "${DOCKERHUB_TOKEN:?Set DOCKERHUB_TOKEN}"
: "${IMAGE_TAG:?Set IMAGE_TAG}"

IMAGE="$DOCKERHUB_USERNAME/much-to-do-backend"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/../backend"

echo "=== Building Docker image: $IMAGE:$IMAGE_TAG ==="
docker build -t "$IMAGE:$IMAGE_TAG" -t "$IMAGE:latest" .

echo ""
echo "=== Pushing to Docker Hub ==="
echo "$DOCKERHUB_TOKEN" | docker login -u "$DOCKERHUB_USERNAME" --password-stdin
docker push "$IMAGE:$IMAGE_TAG"
docker push "$IMAGE:latest"

echo ""
echo "=== Starting ASG rolling instance refresh ==="
aws autoscaling start-instance-refresh \
  --auto-scaling-group-name starttech-asg \
  --strategy Rolling \
  --preferences '{"MinHealthyPercentage": 50, "InstanceWarmup": 120}'

echo ""
echo "=== Waiting for refresh to complete ==="
for i in $(seq 1 20); do
  STATUS=$(aws autoscaling describe-instance-refreshes \
    --auto-scaling-group-name starttech-asg \
    --query "InstanceRefreshes[0].Status" \
    --output text)
  echo "[$i/20] Status: $STATUS"
  [ "$STATUS" = "Successful" ] && echo "Deploy complete!" && exit 0
  [ "$STATUS" = "Failed" ] || [ "$STATUS" = "Cancelled" ] && echo "Deploy failed!" && exit 1
  sleep 30
done
echo "Timeout waiting for deploy"
exit 1
