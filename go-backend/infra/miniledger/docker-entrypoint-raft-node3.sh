#!/bin/sh
set -e
cd /app
npm install --omit=dev
if [ ! -e dashboard ]; then
  ln -sf "$(pwd)/node_modules/miniledger/dashboard" dashboard
fi
if [ ! -d data/node3 ]; then
  echo "[raft-3] init..."
  npx miniledger init -d ./data/node3
fi
MARKER=data/node3/.joined
if [ ! -f "$MARKER" ]; then
  echo "[raft-3] joining cluster via miniledger-1:24440..."
  npx miniledger join "ws://miniledger-1:24440" -d ./data/node3 --p2p-port 24444 --api-port 24445
  touch "$MARKER"
else
  echo "[raft-3] restart start :24445"
  exec npx miniledger start -d ./data/node3 --p2p-port 24444 --api-port 24445
fi
