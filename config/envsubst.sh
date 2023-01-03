#!/bin/bash
set -Eeuo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
cd "$SCRIPT_DIR" || exit 1

# export all variables
set -a

# shellcheck disable=SC1091
source .env
set +a

# create the settings folder to be used for krakend flexible configuration
mkdir -p settings

while IFS= read -r -d '' file; do
  envsubst <"$file" >tmpfile && mv tmpfile "${file/-env/}"
done < <(find ./settings-env -type f -print0)
