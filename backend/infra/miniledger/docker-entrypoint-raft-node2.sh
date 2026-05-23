#!/bin/sh
set -e
cd /app
npm install --omit=dev
if [ ! -e dashboard ]; then
  ln -sf "$(pwd)/node_modules/miniledger/dashboard" dashboard
fi
if [ ! -d data/node2 ]; then
  echo "[raft-2] init..."
  npx miniledger init -d ./data/node2
fi
MARKER=data/node2/.joined
if [ ! -f "$MARKER" ]; then
  echo "[raft-2] joining cluster via miniledger-1:24440..."
  npx miniledger join "ws://miniledger-1:24440" -d ./data/node2 --p2p-port 24442 --api-port 24443
  touch "$MARKER"
else
  echo "[raft-2] restart start :24443"
  exec npx miniledger start -d ./data/node2 --p2p-port 24442 --api-port 24443
fi
