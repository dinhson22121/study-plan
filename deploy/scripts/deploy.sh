#!/bin/sh
# deploy.sh — build/pull images, run migrations, bring up the prod stack, smoke-test.
#
# Idempotent: safe to re-run. Compose reconciles to the desired state; the
# migrate job is a one-shot that no-ops when the schema is already current.
#
# Run from the repo root, or set REPO_ROOT. Requires Docker + Compose v2.
#
# Required env (provide via .env in REPO_ROOT or export beforehand):
#   EDU_JWT_SECRET, EDU_POSTGRES_URL, EDU_S3_ACCESS_KEY, EDU_S3_SECRET_KEY,
#   POSTGRES_PASSWORD  (see deploy/docker-compose.prod.yml + deploy/README.md)
#
# Optional env:
#   IMAGE_TAG        image tag to deploy (default: latest). Set for rollbacks.
#   PULL=1           docker compose pull before up (use with a registry).
#   SKIP_BUILD=1     skip local build (use when pulling prebuilt images).
set -eu

# `set -o pipefail` is not POSIX but supported by bash/dash-on-some-distros;
# enable it when available so a failed stage in a pipe is not masked.
# shellcheck disable=SC3040
(set -o pipefail 2>/dev/null) && set -o pipefail || true

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
REPO_ROOT=${REPO_ROOT:-$(CDPATH= cd -- "$SCRIPT_DIR/../.." && pwd)}
cd "$REPO_ROOT"

BASE_COMPOSE="docker-compose.yml"
PROD_COMPOSE="deploy/docker-compose.prod.yml"
IMAGE_TAG=${IMAGE_TAG:-latest}

export IMAGE_TAG

compose() {
    docker compose -f "$BASE_COMPOSE" -f "$PROD_COMPOSE" "$@"
}

log() {
    printf '\n[deploy] %s\n' "$1"
}

log "Deploying Edu App (IMAGE_TAG=$IMAGE_TAG, root=$REPO_ROOT)"

# 1. Acquire images: pull from registry, or build locally.
if [ "${PULL:-0}" = "1" ]; then
    log "Pulling images from registry"
    compose pull
fi
if [ "${SKIP_BUILD:-0}" != "1" ]; then
    log "Building images"
    compose build
fi

# 2. Run one-shot migrations (exits 0 on success; idempotent).
log "Running database migrations"
compose run --rm migrate

# 3. Bring up the full stack with the prod override.
log "Starting stack"
compose up -d --remove-orphans

# 4. Smoke test (fails the deploy on non-200 health).
log "Running smoke test"
sh "$SCRIPT_DIR/smoke-test.sh"

log "Deploy complete."
