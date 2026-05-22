#!/bin/sh
set -e
cd /app
npm install --omit=dev
if [ ! -d data/node1 ]; then
  echo "[miniledger] initializing node..."
  npm run init
fi
echo "[miniledger] starting node on :4441..."
exec npm run start
