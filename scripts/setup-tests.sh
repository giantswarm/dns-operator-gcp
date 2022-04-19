#!/bin/bash

set -euo pipefail

readonly SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
readonly REPO_ROOT="${SCRIPT_DIR}/.."
readonly ENV_FILE="${REPO_ROOT}/tests/.env"
readonly TEMP_CREDENTIALS_FILE="$(mktemp)"
readonly B64_GOOGLE_APPLICATION_CREDENTIALS="${B64_GOOGLE_APPLICATION_CREDENTIALS:?Base64 encoded GCP credentials not exported}"

echo $B64_GOOGLE_APPLICATION_CREDENTIALS | base64 -d >"$TEMP_CREDENTIALS_FILE"

cat <<EOF >"$ENV_FILE"
GOOGLE_APPLICATION_CREDENTIALS=$TEMP_CREDENTIALS_FILE
EOF
