#!/usr/bin/env bash
# End-to-end flow:
#   1. start infra containers (mongo, per-service redis, kafka)
#      using deployments/local/docker-compose.infra.yml
#   2. wait for mongo to be ready, then apply indexes + seed
#   3. start every Go service as a native host process (one per service)
#   4. wait for the API gateway on :5000
#   5. run the hurl e2e suite in tests/hurl
#   6. stop Go services and tear down infra (volumes optional via E2E_CLEAN=1)
#
# Env knobs:
#   E2E_CLEAN=1   : `docker compose down -v` at the end (drops DB/redis state)
#   E2E_KEEP=1    : skip the teardown at the end (debug/log inspection)
#   E2E_SKIP_RUN=1: assume services already up, only run hurl (then exit)
#   E2E_SKIP_INFRA=1: assume infra already up, just run services + tests
#   E2E_SKIP_TESTS=1: stop after services are up (manual probing)
#   E2E_HURL_RUNS=N: pass-through to hurl/run_tests.sh
#   BASE_URL      : default http://localhost:5000
set -euo pipefail

cd "$(dirname "$0")/.."
ROOT="$(pwd)"

INFRA_COMPOSE="deployments/local/docker-compose.infra.yml"
LOCAL_ENV="deployments/local/local.env"

LOG_DIR="${LOG_DIR:-$ROOT/tmp/e2e-logs}"
PID_DIR="${PID_DIR:-$ROOT/tmp/e2e-pids}"
mkdir -p "$LOG_DIR" "$PID_DIR"

GRPC_SERVICES=(auth role user card merchant saldo topup transaction transfer withdraw)
EXTRA_SERVICES=(email apigateway)
ALL_SERVICES=("${GRPC_SERVICES[@]}" "${EXTRA_SERVICES[@]}")

BASE_URL="${BASE_URL:-http://localhost:5000}"
HURL_RUNS="${E2E_HURL_RUNS:-1}"

C_RED=$'\033[0;31m'; C_GRN=$'\033[0;32m'; C_YEL=$'\033[1;33m'; C_BLU=$'\033[0;34m'; C_RST=$'\033[0m'
say()  { printf '%s[e2e]%s %s\n' "$C_BLU" "$C_RST" "$*"; }
ok()   { printf '%s[ok]%s  %s\n' "$C_GRN" "$C_RST" "$*"; }
warn() { printf '%s[warn]%s %s\n' "$C_YEL" "$C_RST" "$*"; }
die()  { printf '%s[fail]%s %s\n' "$C_RED" "$C_RST" "$*" >&2; exit 1; }

cleanup_services() {
  say "Stopping Go services..."
  if [[ -d "$PID_DIR" ]]; then
    for pidfile in "$PID_DIR"/*.pid; do
      [[ -e "$pidfile" ]] || continue
      local pid
      pid="$(cat "$pidfile" 2>/dev/null || true)"
      if [[ -n "$pid" ]] && kill -0 "$pid" 2>/dev/null; then
        kill "$pid" 2>/dev/null || true
      fi
      rm -f "$pidfile"
    done
  fi
  # Best-effort reap any stragglers from a prior crashed run.
  pkill -f 'service/(auth|role|user|card|merchant|saldo|topup|transaction|transfer|withdraw|email|apigateway)/cmd/main.go' 2>/dev/null || true
}

cleanup_infra() {
  if [[ "${E2E_KEEP:-0}" == "1" ]]; then
    warn "E2E_KEEP=1; skipping infra teardown"
    return
  fi
  say "Tearing down infra containers..."
  local down_args=(-f "$INFRA_COMPOSE" down --remove-orphans)
  if [[ "${E2E_CLEAN:-0}" == "1" ]]; then
    down_args+=(-v)
  fi
  docker compose "${down_args[@]}"
}

trap 'rc=$?; cleanup_services || true; if [[ "${E2E_SKIP_INFRA:-0}" != "1" ]]; then cleanup_infra || true; fi; exit $rc' EXIT INT TERM

require() {
  command -v "$1" >/dev/null 2>&1 || die "missing dependency: $1"
}

require docker
require go
require hurl
# `docker compose` (v2) is the only form this script uses.
docker compose version >/dev/null 2>&1 || die "missing 'docker compose' v2 plugin"

# ---------------------------------------------------------------------------
# 1. infra up
# ---------------------------------------------------------------------------
if [[ "${E2E_SKIP_INFRA:-0}" != "1" ]]; then
  say "Starting infra from $INFRA_COMPOSE"
  docker compose -f "$INFRA_COMPOSE" up -d
else
  say "E2E_SKIP_INFRA=1; assuming infra is already running"
fi

# ---------------------------------------------------------------------------
# 2. wait for mongo, apply indexes + seed
# ---------------------------------------------------------------------------
wait_tcp() {
  local host="$1" port="$2" tries="${3:-60}"
  for ((i = 1; i <= tries; i++)); do
    if (exec 3<>"/dev/tcp/$host/$port") 2>/dev/null; then
      exec 3>&- 3<&-
      return 0
    fi
    sleep 1
  done
  return 1
}

wait_healthy() {
  local svc="$1" tries="${2:-60}"
  for ((i = 1; i <= tries; i++)); do
    local cid
    cid="$(docker compose -f "$INFRA_COMPOSE" ps -q "$svc" 2>/dev/null || true)"
    if [[ -n "$cid" ]] && [[ "$(docker inspect -f '{{.State.Health.Status}}' "$cid" 2>/dev/null || echo starting)" == "healthy" ]]; then
      return 0
    fi
    sleep 1
  done
  return 1
}

say "Waiting for mongo on localhost:27017 ..."
wait_tcp localhost 27017 90 || die "mongo never came up on :27017"
wait_healthy mongo 60 || warn "mongo healthcheck did not flip to healthy; continuing"
ok "mongo is reachable"

say "Loading env from $LOCAL_ENV"
set -a
# shellcheck disable=SC1090
source "$LOCAL_ENV"
set +a

say "Applying MongoDB collections and indexes — retrying up to 3 times"
MIGRATE_OK=0
for ((attempt = 1; attempt <= 3; attempt++)); do
  if go run ./service/migrate/main.go up; then
    MIGRATE_OK=1
    break
  fi
  warn "migrate attempt $attempt failed; retrying in 3s"
  sleep 3
done
[[ "$MIGRATE_OK" == "1" ]] || die "migrate failed after 3 attempts"

USER_COUNT="$(docker exec mongo mongosh --quiet -u DRAGON -p DRAGON --authenticationDatabase admin PAYMENT_GATEWAY --eval "db.users.countDocuments({})" 2>/dev/null | tr -d ' ' || echo 0)"
if [[ "${USER_COUNT:-0}" == "0" ]]; then
  say "Seeding database (mongo seeder)"
  go run ./service/migrate/main.go seed
  ok "seeded"
else
  say "users collection already populated ($USER_COUNT docs); skipping seed"
fi

# ---------------------------------------------------------------------------
# 3. start Go services natively
# ---------------------------------------------------------------------------
if [[ "${E2E_SKIP_RUN:-0}" != "1" ]]; then
  cleanup_services
  for svc in "${ALL_SERVICES[@]}"; do
    svc_log="$LOG_DIR/$svc.log"
    say "Starting $svc (log: $svc_log)"
    (
      cd "$ROOT/service/$svc" || die "service dir not found: service/$svc"
      nohup go run ./cmd/main.go >"$svc_log" 2>&1 &
      echo $! > "$PID_DIR/$svc.pid"
    )
  done
else
  say "E2E_SKIP_RUN=1; assuming Go services are already up"
fi

# ---------------------------------------------------------------------------
# 4. wait for api gateway
# ---------------------------------------------------------------------------
say "Waiting for API Gateway on $BASE_URL/health ..."
GATEWAY_READY=0
for ((i = 1; i <= 120; i++)); do
  if curl -fsS -o /dev/null "$BASE_URL/health" 2>/dev/null; then
    GATEWAY_READY=1
    break
  fi
  sleep 1
done
if [[ "$GATEWAY_READY" != "1" ]]; then
  die "API Gateway did not become ready in time. Tail: tail -n 100 $LOG_DIR/apigateway.log"
fi
ok "API Gateway is up at $BASE_URL"

# ---------------------------------------------------------------------------
# 5. hurl e2e
# ---------------------------------------------------------------------------
if [[ "${E2E_SKIP_TESTS:-0}" == "1" ]]; then
  say "E2E_SKIP_TESTS=1; services are up. Press Ctrl-C to stop. (PIDs: $(cat "$PID_DIR"/*.pid | tr '\n' ' '))"
  wait
  exit 0
fi

say "Running hurl e2e suite against $BASE_URL (runs=$HURL_RUNS)"
set +e
HURL_RUNS="$HURL_RUNS" BASE_URL="$BASE_URL" bash hurl/run_tests.sh
HURL_RC=$?
set -e

if [[ "$HURL_RC" -ne 0 ]]; then
  warn "hurl suite returned $HURL_RC; logs in $LOG_DIR"
  exit "$HURL_RC"
fi

ok "hurl e2e passed"
exit 0
