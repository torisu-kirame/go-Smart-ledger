package com.smartledger.gateway.config;

import com.smartledger.common.jwt.JwtService;
import org.springframework.cloud.gateway.filter.GatewayFilterChain;
import org.springframework.cloud.gateway.filter.GlobalFilter;
import org.springframework.core.Ordered;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.MediaType;
import org.springframework.stereotype.Component;
import org.springframework.web.server.ServerWebExchange;
import reactor.core.publisher.Mono;

import java.nio.charset.StandardCharsets;
import java.util.Set;

/**
 * 网关 JWT 校验：解析 Bearer 令牌，向下游注入用户身份头。
 * <p>
 * 下游服务通过 {@code X-User-Id} / {@code X-Username} 获取当前用户，无需重复验签。
 */
@Component
public class JwtAuthFilter implements GlobalFilter, Ordered {

    /** 无需 Bearer 的公开路径 */
    private static final Set<String> PUBLIC_PREFIXES = Set.of(
            "/api/v1/auth/captcha",
            "/api/v1/auth/login",
            "/api/v1/auth/register",
            "/api/v1/auth/refresh",
            "/api/v1/auth/logout",
            "/api/v1/auth/health"
    );

    private final JwtService jwtService;

    public JwtAuthFilter(JwtService jwtService) {
        this.jwtService = jwtService;
    }

    @Override
    public Mono<Void> filter(ServerWebExchange exchange, GatewayFilterChain chain) {
        String path = exchange.getRequest().getURI().getPath();
        if ("/api/v1/health".equals(path) || isPublic(path)) {
            return chain.filter(exchange);
        }
        if (!path.startsWith("/api/v1/")) {
            return chain.filter(exchange);
        }
        String auth = exchange.getRequest().getHeaders().getFirst(HttpHeaders.AUTHORIZATION);
        if (auth == null || !auth.startsWith("Bearer ")) {
            return unauthorized(exchange, "missing access token");
        }
        String token = auth.substring("Bearer ".length()).trim();
        try {
            var user = jwtService.parseAccess(token);
            var mutated = exchange.getRequest().mutate()
                    .header("X-User-Id", user.userId())
                    .header("X-Username", user.username())
                    .build();
            return chain.filter(exchange.mutate().request(mutated).build());
        } catch (Exception e) {
            return unauthorized(exchange, "invalid or expired access token");
        }
    }

    private static boolean isPublic(String path) {
        for (String p : PUBLIC_PREFIXES) {
            if (path.equals(p) || path.startsWith(p + "/")) {
                return true;
            }
        }
        return false;
    }

    private static Mono<Void> unauthorized(ServerWebExchange exchange, String msg) {
        exchange.getResponse().setStatusCode(HttpStatus.UNAUTHORIZED);
        exchange.getResponse().getHeaders().setContentType(MediaType.APPLICATION_JSON);
        byte[] body = ("{\"code\":401,\"msg\":\"" + msg + "\"}").getBytes(StandardCharsets.UTF_8);
        return exchange.getResponse().writeWith(Mono.just(exchange.getResponse().bufferFactory().wrap(body)));
    }

    @Override
    public int getOrder() {
        return -100;
    }
}
