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
if [ -f .env ]; then
    docker container run -d \
        --name "$CONTAINER_NAME" \
        -p "$PORT:$PORT" \
        --env-file .env \
        "$IMAGE_NAME"
else
    docker container run -d \
        --name "$CONTAINER_NAME" \
        -p "$PORT:$PORT" \
        "$IMAGE_NAME"
fi

echo "Forum is running at http://localhost:$PORT"
