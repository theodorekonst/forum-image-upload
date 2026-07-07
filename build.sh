#!/bin/bash
set -e

IMAGE_NAME="forum"
CONTAINER_NAME="forum"
PORT="8080"

echo "Building Docker image..."
docker image build -t "$IMAGE_NAME" .

echo "Stopping existing container (if any)..."
docker rm -f "$CONTAINER_NAME" 2>/dev/null || true

echo "Starting container..."
docker container run -d \
    --name "$CONTAINER_NAME" \
    -p "$PORT:$PORT" \
    "$IMAGE_NAME"

echo "Forum is running at http://localhost:$PORT"
