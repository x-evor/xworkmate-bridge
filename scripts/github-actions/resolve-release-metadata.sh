#!/usr/bin/env bash
set -euo pipefail

short_sha="${GITHUB_SHA::7}"
tag="main-${short_sha}"
title="main ${short_sha}"

printf 'tag=%s\n' "${tag}" >> "${GITHUB_OUTPUT}"
printf 'title=%s\n' "${title}" >> "${GITHUB_OUTPUT}"
