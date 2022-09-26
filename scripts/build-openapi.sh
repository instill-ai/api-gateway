#!/bin/bash
set -Eeuo pipefail

# Ensure we start in the parent directory of where this script is (the root folder).
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
ROOT_DIR="$SCRIPT_DIR/.."
cd "$ROOT_DIR" || exit 1

SERVER_URL="https://api.instill.tech"

# the temp directory used, within $DIR
# omit the -p parameter to create a temporal directory in the default location
WORK_DIR=$(mktemp -d 2>/dev/null || mktemp -d -t "$SCRIPT_DIR")

# check if tmp dir was created
if [[ ! "$WORK_DIR" || ! -d "$WORK_DIR" ]]; then
  echo "Could not create temp dir!"
  exit 1
fi

# Generate config file and API spec for OpenAPI generator
PUBLIC_FULL_VERSION=$(cat ./version-public.txt)
IFS=" " read -r -a array <<<"${PUBLIC_FULL_VERSION//./ }" # replace points, split into array
if [ "${#array[@]}" -lt "3" ]; then
    echo -e "Public version $PUBLIC_FULL_VERSION is not in semantic versioning format: X.Y.Z"
    exit 1
fi

PUBLIC_VERSION="${array[0]}.${array[1]}" # extract X.Y and fill into OpenAPI yaml file
VERSION="${PUBLIC_VERSION}" SERVER_URL="$SERVER_URL" mo api-gateway/api/openapi.yaml.mustache > "$WORK_DIR/api-gateway-openapi.yaml"

echo "Fetch the latest OpenAPI spec from each backend..."
OPENAPI_MANAGEMENT_GCS_URL=$(gsutil ls 'gs://public-europe-west2-c-artifacts/docs/api/management/openapi_*.yaml' | sort -V | tail -n 1)
OPENAPI_MODEL_GCS_URL=$(gsutil ls 'gs://public-europe-west2-c-artifacts/docs/api/model/openapi_*.yaml' | sort -V | tail -n 1)
OPENAPI_PIPELINE_GCS_URL=$(gsutil ls 'gs://public-europe-west2-c-artifacts/docs/api/pipeline/openapi_*.yaml' | sort -V | tail -n 1)

OPENAPI_MANAGEMENT_PATH="management/$(basename "$OPENAPI_MANAGEMENT_GCS_URL")"
OPENAPI_MODEL_PATH="model/$(basename "$OPENAPI_MODEL_GCS_URL")"
OPENAPI_PIPELINE_PATH="pipeline/$(basename "$OPENAPI_PIPELINE_GCS_URL")"

OPENAPI_MANAGEMENT_URL="https://artifacts.instill.tech/docs/api/$OPENAPI_MANAGEMENT_PATH" \
OPENAPI_MODEL_URL="https://artifacts.instill.tech/docs/api/$OPENAPI_MODEL_PATH" \
OPENAPI_PIPELINE_URL="https://artifacts.instill.tech/docs/api/$OPENAPI_PIPELINE_PATH" \
mo api-gateway/api/openapi-merge.json.mustache > "$WORK_DIR"/openapi-merge.json

echo "Merge OpenAPI files..."
npx openapi-merge-cli -c "$WORK_DIR"/openapi-merge.json

cp "$WORK_DIR/openapi.yaml" "$ROOT_DIR/api/openapi.yaml"

# re-write merge log file
echo -e "Merged OpenAPI public version: v$PUBLIC_FULL_VERSION\n- $OPENAPI_MANAGEMENT_PATH\n- $OPENAPI_MODEL_PATH\n- $OPENAPI_PIPELINE_PATH" > "$ROOT_DIR/api/build-openapi.log"
echo "Finished merging openapi with v$PUBLIC_FULL_VERSION and writing to $ROOT_DIR/api/build-openapi.log"

# deletes the temp directory
function cleanup {
  rm -rf "$WORK_DIR"
  echo "Deleted temp working directory $WORK_DIR"
}

# register the cleanup function to be called on the EXIT signal
trap cleanup EXIT
