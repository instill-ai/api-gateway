#!/bin/bash
set -Eeuo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
cd "$SCRIPT_DIR" || exit 1

# export all variables
set -a

# shellcheck disable=SC1091
source .env
set +a

# Derive non-headless HTTP hosts from (possibly headless) gRPC hosts.
# Helm sets *_BACKEND_HOST to the headless Service for gRPC dns:/// LB;
# HTTP/1.1 backends need the regular ClusterIP Service so kube-proxy DNAT
# distributes per-connection.  Stripping "-headless" is a no-op for Docker
# Compose where the suffix is absent.
set -a
MGMT_BACKEND_HTTP_HOST="${MGMT_BACKEND_HOST%-headless}"
MODEL_BACKEND_HTTP_HOST="${MODEL_BACKEND_HOST%-headless}"
PIPELINE_BACKEND_HTTP_HOST="${PIPELINE_BACKEND_HOST%-headless}"
ARTIFACT_BACKEND_HTTP_HOST="${ARTIFACT_BACKEND_HOST%-headless}"
set +a

# create the settings folder to be used for krakend flexible configuration
mkdir -p settings

while IFS= read -r -d '' file; do
  envsubst <"$file" >tmpfile && mv tmpfile ./settings/"$(basename -- "${file}")"
done < <(find ./share/settings-env -type f -print0)

while IFS= read -r -d '' file; do
  envsubst <"$file" >tmpfile && mv tmpfile ./settings/"$(basename -- "${file}")"
done < <(find ./settings-env -type f -print0)
