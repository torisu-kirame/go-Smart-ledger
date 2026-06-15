#!/usr/bin/env bash
set -euo pipefail

DATA="/data"
NODE_ROOT="${DATA}/nodes/127.0.0.1"
NODE_DIR="${NODE_ROOT}/node0"
BIN="${NODE_ROOT}/fisco-bcos"

BUILD_CHAIN_URL="${FISCO_BUILD_CHAIN_URL:-https://github.com/FISCO-BCOS/FISCO-BCOS/releases/download/v3.6.0/build_chain.sh}"
BUILD_CHAIN_MASTER="${FISCO_BUILD_CHAIN_MASTER:-https://raw.githubusercontent.com/FISCO-BCOS/FISCO-BCOS/master/tools/BcosAirBuilder/build_chain.sh}"
BUILD_CHAIN_MIRROR="${FISCO_BUILD_CHAIN_MIRROR:-https://osp-1257653870.cos.ap-guangzhou.myqcloud.com/FISCO-BCOS/FISCO-BCOS/releases/v3.6.0/build_chain.sh}"
TASSL_GITHUB_URL="${FISCO_TASSL_URL:-https://github.com/FISCO-BCOS/TASSL/releases/download/V_1.4/tassl-1.1.1b-linux-x86_64.tar.gz}"

fetch_build_chain() {
  local dst="${DATA}/build_chain.sh"
  if curl -fsSL "${BUILD_CHAIN_MASTER}" -o "${dst}"; then
    return 0
  fi
  if curl -fsSL "${BUILD_CHAIN_URL}" -o "${dst}"; then
    return 0
  fi
  echo "[fisco] GitHub build_chain unavailable, trying mirror..."
  curl -fsSL "${BUILD_CHAIN_MIRROR}" -o "${dst}"
}

install_tassl() {
  local tassl_bin="${HOME}/.fisco/tassl-1.1.1b"
  if [ -x "${tassl_bin}" ]; then
    return 0
  fi
  echo "[fisco] installing tassl from GitHub (CDN mirror often returns 403)..."
  mkdir -p "${HOME}/.fisco"
  local tmp="${DATA}/tassl-1.1.1b-linux-x86_64.tar.gz"
  curl -fsSL "${TASSL_GITHUB_URL}" -o "${tmp}"
  tar -zxf "${tmp}" -C "${DATA}"
  chmod u+x "${DATA}/tassl-1.1.1b-linux-x86_64"
  mv "${DATA}/tassl-1.1.1b-linux-x86_64" "${tassl_bin}"
  rm -f "${tmp}"
}

patch_node_ini() {
  local ini="${NODE_DIR}/config.ini"
  sed -i 's/listen_ip=127.0.0.1/listen_ip=0.0.0.0/g' "${ini}" || true
  sed -i 's/;disable_ssl=true/disable_ssl=true/g' "${ini}" || true
  sed -i 's/disable_ssl=false/disable_ssl=true/g' "${ini}" || true
}

init_chain() {
  echo "[fisco] first run — downloading build_chain.sh and generating single-node chain (group0)..."
  mkdir -p "${DATA}"
  cd "${DATA}"
  rm -rf nodes build_chain.sh
  install_tassl
  fetch_build_chain
  chmod +x build_chain.sh
  bash build_chain.sh -l 127.0.0.1:1 -p 30300,20200 -v v3.6.0
  patch_node_ini
  if [ -f nodes/build.log ]; then
    grep -Ei "Admin account" nodes/build.log > admin.txt || true
    cp -f nodes/build.log build.log
  fi
  rm -f build_chain.sh
  echo "[fisco] chain ready. RPC http://0.0.0.0:20200 (group0 / chain0)"
  if [ -f admin.txt ]; then
    cat admin.txt
  fi
  echo "[fisco] deploy LedgerRegistry.sol then set RegistryContract + PrivateKeyHex on ledger-api"
}

if [ ! -f "${NODE_DIR}/config.ini" ]; then
  if [ -d "${DATA}/nodes" ]; then
    echo "[fisco] removing incomplete chain data from a previous failed init..."
    rm -rf "${DATA}/nodes"
  fi
  init_chain
fi

if [ ! -x "${BIN}" ]; then
  echo "[fisco] ERROR: missing executable ${BIN}" >&2
  exit 1
fi

echo "[fisco] starting node0..."
cd "${NODE_DIR}"
exec "${BIN}" -c config.ini -g config.genesis
