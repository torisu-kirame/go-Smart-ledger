package com.smartledger.gateway.config;

import org.springframework.boot.context.properties.EnableConfigurationProperties;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.cloud.gateway.route.RouteLocator;
import org.springframework.cloud.gateway.route.builder.RouteLocatorBuilder;

/**
 * 按业务域将请求转发至 auth / ledger / storage 微服务。
 * <p>
 * {@code /api/v1/health} 不在此注册，由 {@link com.smartledger.gateway.web.HealthController} 本地聚合。
 */
@Configuration
@EnableConfigurationProperties(UpstreamProperties.class)
public class GatewayRoutesConfig {

    @Bean
    RouteLocator routes(RouteLocatorBuilder builder, UpstreamProperties upstreams) {
        return builder.routes()
                .route("auth-api", r -> r.path(
                                "/api/v1/auth/**",
                                "/api/v1/users/**",
                                "/api/v1/teams/**",
                                "/api/v1/entry-templates/**",
                                "/api/v1/friends/**")
                        .uri(upstreams.auth()))
                .route("ledger-api", r -> r.path(
                                "/api/v1/ledgers/**",
                                "/api/v1/chain/**",
                                "/api/v1/entry-schema/**",
                                "/api/v1/import/**")
                        .uri(upstreams.ledger()))
                .route("ai-api", r -> r.path("/api/v1/ai/**")
                        .uri(upstreams.ai()))
                .route("storage-api", r -> r.path("/api/v1/storage/**")
                        .uri(upstreams.storage()))
                .build();
    }
}
