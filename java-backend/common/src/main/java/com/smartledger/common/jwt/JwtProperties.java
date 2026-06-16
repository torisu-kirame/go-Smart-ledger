package com.smartledger.common.jwt;

import org.springframework.boot.context.properties.ConfigurationProperties;

/** JWT 密钥与过期时间，可通过 application.yml 或环境变量覆盖 */
@ConfigurationProperties(prefix = "smartledger.jwt")
public record JwtProperties(
        String accessSecret,
        String refreshSecret,
        long accessExpireSeconds,
        long refreshExpireSeconds,
        boolean cookieSecure
) {
    public JwtProperties {
        if (accessSecret == null || accessSecret.isBlank()) {
            accessSecret = "smart-ledger-access-secret-change-me";
        }
        if (refreshSecret == null || refreshSecret.isBlank()) {
            refreshSecret = "smart-ledger-refresh-secret-change-me";
        }
        if (accessExpireSeconds <= 0) {
            accessExpireSeconds = 900;
        }
        if (refreshExpireSeconds <= 0) {
            refreshExpireSeconds = 604800;
        }
    }
}
