#!/usr/bin/env bash
set -euo pipefail

ARTIFACT_DIR="${1:-dist}"
mkdir -p "${ARTIFACT_DIR}"

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o "${ARTIFACT_DIR}/xworkmate-bridge" .
