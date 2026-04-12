#!/usr/bin/env bash
set -euo pipefail

image_repo="${SERVICE_REGISTRY:?SERVICE_REGISTRY is required}/${SERVICE_IMAGE_REPO_OWNER:?SERVICE_IMAGE_REPO_OWNER is required}/${SERVICE_IMAGE_NAME:?SERVICE_IMAGE_NAME is required}"
image_tag="${GITHUB_SHA:?GITHUB_SHA is required}"
image_ref="${image_repo}:${image_tag}"

printf 'image_repo=%s\n' "${image_repo}" >> "${GITHUB_OUTPUT}"
printf 'image_tag=%s\n' "${image_tag}" >> "${GITHUB_OUTPUT}"
printf 'image_ref=%s\n' "${image_ref}" >> "${GITHUB_OUTPUT}"
