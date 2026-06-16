package com.smartledger.gateway.config;

import org.springframework.boot.context.properties.ConfigurationProperties;

/** Compose 内网服务地址，可通过 application.yml 覆盖 */
@ConfigurationProperties(prefix = "smartledger.upstreams")
public record UpstreamProperties(String auth, String ledger, String storage) {
    public UpstreamProperties {
        if (auth == null || auth.isBlank()) auth = "http://auth-api:28887";
        if (ledger == null || ledger.isBlank()) ledger = "http://ledger-api:28888";
        if (storage == null || storage.isBlank()) storage = "http://storage-api:28890";
    }
}
