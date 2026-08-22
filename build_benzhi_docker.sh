#!/bin/bash
set -euo pipefail

image_name=${1:-deformation-network}
platform=${2:-linux/amd64}

docker buildx build --platform "$platform" --load -f benzhi.Dockerfile -t "$image_name" .
docker run --rm "$image_name" --smoke-test
