#!/bin/sh
set -e
cd /app
npm install --omit=dev
# miniledger createApp() looks for ./dashboard under cwd; npm 包内路径为 node_modules/miniledger/dashboard
if [ ! -e dashboard ]; then
  ln -sf "$(pwd)/node_modules/miniledger/dashboard" dashboard
fi
if [ ! -d data/node1 ]; then
  echo "[miniledger] initializing node..."
  npm run init
fi
echo "[miniledger] starting node on :24441..."
exec npm run start
