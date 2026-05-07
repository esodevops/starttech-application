#!/bin/bash
# -------------------------------------------------
# rollback.sh
#
# Rolls the backend back to a previous Docker image
# tag by updating the ASG launch template and
# triggering an instance refresh.
#
# Usage:
#   ./scripts/rollback.sh v1.0.0
# -------------------------------------------------

set -e

ROLLBACK_TAG="${1:?Usage: ./rollback.sh <image-tag>}"
DOCKERHUB_USERNAME="${DOCKERHUB_USERNAME:?Set DOCKERHUB_USERNAME}"
IMAGE="$DOCKERHUB_USERNAME/much-to-do-backend:$ROLLBACK_TAG"

echo "=== Rolling back to image: $IMAGE ==="

# Get the current launch template ID
LT_ID=$(aws ec2 describe-launch-templates \
  --filters "Name=tag:Name,Values=starttech-launch-template" \
  --query "LaunchTemplates[0].LaunchTemplateId" \
  --output text)

echo "Launch Template ID: $LT_ID"

# Get the current user_data
CURRENT_USER_DATA=$(aws ec2 describe-launch-template-versions \
  --launch-template-id "$LT_ID" --versions '$Latest' \
  --query "LaunchTemplateVersions[0].LaunchTemplateData.UserData" \
  --output text | base64 -d)

# Replace the image reference with the rollback tag
NEW_USER_DATA=$(echo "$CURRENT_USER_DATA" | \
  sed "s|$DOCKERHUB_USERNAME/much-to-do-backend:.*|$IMAGE|g")

# Create a new launch template version with the rollback image
aws ec2 create-launch-template-version \
  --launch-template-id "$LT_ID" \
  --source-version '$Latest' \
  --launch-template-data "{\"UserData\":\"$(echo "$NEW_USER_DATA" | base64 -w0)\"}"

echo ""
echo "=== Starting rollback instance refresh ==="
aws autoscaling start-instance-refresh \
  --auto-scaling-group-name starttech-asg \
  --strategy Rolling \
  --preferences '{"MinHealthyPercentage": 50, "InstanceWarmup": 120}'

echo "Rollback started. Monitor in AWS Console → EC2 → Auto Scaling Groups → starttech-asg"
