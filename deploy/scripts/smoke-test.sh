#!/bin/sh
# smoke-test.sh — post-deploy health verification for the Edu App.
#
# Checks (exit non-zero on any failure):
#   1. GET /health         must be 200      (liveness)
#   2. GET /health/ready   must be 200      (Postgres + Redis readiness)
#   3. GET /metrics        must be reachable internally (Prometheus)
#   4. (optional) register + login round-trip against /api/v1/auth
#
# Configuration via env:
#   BASE_URL         public base, default https://localhost (nginx, -k for self-signed)
#   INTERNAL_URL     in-network base for /metrics, default http://app:8080
#                    (when run on the host outside compose, override e.g.
#                     INTERNAL_URL=http://localhost:8080 with a temp port-forward)
#   RUN_AUTH_TEST=1  also exercise register+login (uses a throwaway email)
#   CURL_OPTS        extra curl flags (e.g. "-k" to accept self-signed TLS)
set -eu

BASE_URL=${BASE_URL:-https://localhost}
INTERNAL_URL=${INTERNAL_URL:-http://app:8080}
# Accept self-signed certs by default for the local/staging smoke run.
CURL_OPTS=${CURL_OPTS:--k}

fail() {
    printf '[smoke] FAIL: %s\n' "$1" >&2
    exit 1
}

pass() {
    printf '[smoke] OK: %s\n' "$1"
}

# http_status URL -> prints the numeric status code
http_status() {
    # shellcheck disable=SC2086
    curl $CURL_OPTS -s -o /dev/null -w '%{http_code}' --max-time 10 "$1"
}

# 1. Liveness
code=$(http_status "$BASE_URL/health") || fail "/health request errored"
[ "$code" = "200" ] || fail "/health returned $code (want 200)"
pass "/health -> 200"

# 2. Readiness
code=$(http_status "$BASE_URL/health/ready") || fail "/health/ready request errored"
[ "$code" = "200" ] || fail "/health/ready returned $code (want 200)"
pass "/health/ready -> 200"

# 3. Metrics — reachable on the internal endpoint (public access is denied by nginx).
code=$(http_status "$INTERNAL_URL/metrics") || fail "/metrics request errored (internal)"
[ "$code" = "200" ] || fail "/metrics (internal) returned $code (want 200)"
pass "/metrics (internal) -> 200"

# 4. Optional auth round-trip.
if [ "${RUN_AUTH_TEST:-0}" = "1" ]; then
    # Password policy: 10-72 chars, must contain a letter and a digit.
    EMAIL="smoke+$(date +%s)-$$@example.com"
    PASSWORD="SmokeTest12345"
    BODY="{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}"

    # Register (201/200 acceptable).
    # shellcheck disable=SC2086
    code=$(curl $CURL_OPTS -s -o /dev/null -w '%{http_code}' --max-time 10 \
        -H 'Content-Type: application/json' \
        -d "$BODY" "$BASE_URL/api/v1/auth/register") || fail "register request errored"
    case "$code" in
        200|201) pass "auth/register -> $code" ;;
        *) fail "auth/register returned $code (want 200/201)" ;;
    esac

    # Login.
    LOGIN_BODY="{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}"
    # shellcheck disable=SC2086
    code=$(curl $CURL_OPTS -s -o /dev/null -w '%{http_code}' --max-time 10 \
        -H 'Content-Type: application/json' \
        -d "$LOGIN_BODY" "$BASE_URL/api/v1/auth/login") || fail "login request errored"
    [ "$code" = "200" ] || fail "auth/login returned $code (want 200)"
    pass "auth/login -> 200"
fi

printf '[smoke] all checks passed\n'
