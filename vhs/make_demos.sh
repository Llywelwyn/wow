#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "building wow-vhs docker image"
docker build -t wow-vhs -f "$REPO_ROOT/vhs/Dockerfile" "$REPO_ROOT"
echo "running vhs"
for tape in "$REPO_ROOT"/vhs/*.tape; do
  echo "recording $(basename "$tape")"
  docker run --rm \
    -v "$REPO_ROOT":/repo \
    -e WOW_HOME=/tmp/wow-home \
    --tmpfs /tmp \
    wow-vhs "./vhs/$(basename "$tape")"
done
echo "done, demo images saved in $REPO_ROOT/vhs"
