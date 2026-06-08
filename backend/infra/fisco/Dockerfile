# FISCO BCOS 3.x Air — 单节点（首次启动时 build_chain 初始化）

FROM ubuntu:22.04

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        curl ca-certificates openssl wget procps bash \
    && rm -rf /var/lib/apt/lists/*

COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

WORKDIR /data
VOLUME ["/data"]

EXPOSE 20200 30300

LABEL org.opencontainers.image.title="smart-ledger-fisco-node" \
      org.opencontainers.image.description="FISCO BCOS 3.x Air single-node (init on first run)"

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
