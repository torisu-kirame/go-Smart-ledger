#!/bin/sh
set -e
cd /app
npm install --omit=dev
if [ ! -d data/node1 ]; then
  echo "[raft-1] init..."
  npx miniledger init -d ./data/node1
fi
echo "[raft-1] starting bootstrap node (raft) :24441 / :24440"
exec npx miniledger start -d ./data/node1 --consensus raft --api-port 24441 --p2p-port 24440
