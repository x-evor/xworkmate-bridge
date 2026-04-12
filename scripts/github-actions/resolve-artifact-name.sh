#!/usr/bin/env bash
set -euo pipefail

printf 'artifact_name=xworkmate-bridge-service-image-%s\n' "${GITHUB_SHA:?GITHUB_SHA is required}" >> "${GITHUB_OUTPUT}"
