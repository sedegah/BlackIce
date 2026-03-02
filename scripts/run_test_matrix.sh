#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOG_DIR="$ROOT_DIR/tests/logs"
mkdir -p "$LOG_DIR"

echo "[1/4] go unit+integration tests"
go test ./...

echo "[2/4] python unit tests"
python3 -m unittest discover -s "$ROOT_DIR/ml/inference" -p 'test_*.py'

echo "[3/4] sidecar + runtime smoke"
SOCK="$LOG_DIR/blackice-matrix.sock"
rm -f "$SOCK"
python3 "$ROOT_DIR/ml/inference/server.py" --socket "$SOCK" >"$LOG_DIR/python_smoke.log" 2>&1 &
PY_PID=$!
sleep 1

(cd "$ROOT_DIR" && go run ./cmd/blackice --socket "$SOCK" --pps 120000 --window 1s >"$LOG_DIR/go_smoke.log" 2>&1) &
GO_PID=$!
sleep 3
kill "$GO_PID" >/dev/null 2>&1 || true
kill "$PY_PID" >/dev/null 2>&1 || true
wait "$GO_PID" 2>/dev/null || true
wait "$PY_PID" 2>/dev/null || true

echo "[4/4] quick assertion of mitigation output"
if ! grep -q "mitigation=" "$LOG_DIR/go_smoke.log"; then
  echo "expected mitigation output missing"
  exit 1
fi

echo "matrix complete"
