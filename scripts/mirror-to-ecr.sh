#!/bin/bash
# Mirror external Docker images to ECR
# Run once to internalize all dependencies

set -euo pipefail

ECR_REGISTRY="110380501820.dkr.ecr.us-east-1.amazonaws.com"
ECR_PREFIX="pentagi"

# Login to ECR
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin "$ECR_REGISTRY"

# Images to mirror: source -> ECR repo name
declare -A IMAGES=(
  ["kalilinux/kali-rolling:latest"]="kali-rolling"
  ["pgvector/pgvector:pg16"]="pgvector"
  ["docker:27-dind"]="docker-dind"
  ["neo4j:5.26.2"]="neo4j"
  ["node:23-slim"]="node"
  ["golang:1.24-bookworm"]="golang"
  ["alpine:3.23.3"]="alpine"
  ["browserless/chrome:1.61.1-chrome-stable"]="browserless-chrome"
)

for SOURCE in "${!IMAGES[@]}"; do
  REPO="${IMAGES[$SOURCE]}"
  TAG="${SOURCE##*:}"
  ECR_IMAGE="$ECR_REGISTRY/$ECR_PREFIX/$REPO:$TAG"

  echo "=== Mirroring $SOURCE -> $ECR_IMAGE ==="

  # Create repo if not exists
  aws ecr describe-repositories --repository-names "$ECR_PREFIX/$REPO" --region us-east-1 2>/dev/null || \
    aws ecr create-repository --repository-name "$ECR_PREFIX/$REPO" --region us-east-1

  docker pull --platform linux/amd64 "$SOURCE"
  docker tag "$SOURCE" "$ECR_IMAGE"
  docker push "$ECR_IMAGE"

  echo "=== Done: $ECR_IMAGE ==="
  echo ""
done

echo "All images mirrored to ECR!"
