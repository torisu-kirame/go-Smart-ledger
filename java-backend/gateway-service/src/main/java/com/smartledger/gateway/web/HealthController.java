package com.smartledger.gateway.web;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.reactive.function.client.WebClient;

import java.util.Map;

/**
 * 网关聚合健康检查，供控制台 Dashboard 与 Docker healthcheck 使用。
 * <p>
 * ledger 不可达时不抛错，仅标记 miniLedgerOnline=false。
 */
@RestController
public class HealthController {
    private final WebClient ledger;

    public HealthController(@Value("${smartledger.upstreams.ledger}") String ledgerUrl) {
        this.ledger = WebClient.builder().baseUrl(ledgerUrl).build();
    }

    @GetMapping("/api/v1/health")
    public Map<String, Object> health() {
        boolean miniOnline = false;
        int pending = 0;
        int failed = 0;
        try {
            @SuppressWarnings("unchecked")
            Map<String, Object> body = ledger.get()
                    .uri("/api/v1/health")
                    .retrieve()
                    .bodyToMono(Map.class)
                    .block();
            if (body != null) {
                miniOnline = Boolean.TRUE.equals(body.get("miniLedgerOnline"));
                pending = intVal(body.get("queuePending"));
                failed = intVal(body.get("queueFailed"));
            }
        } catch (Exception ignored) {
            // 下游故障不影响网关自身存活判定
        }
        return Map.of(
                "status", "ok",
                "gateway", "ok",
                "backend", "java",
                "miniLedgerOnline", miniOnline,
                "chainQueuePending", pending,
                "chainQueueFailed", failed
        );
    }

    private static int intVal(Object v) {
        if (v instanceof Number n) {
            return n.intValue();
        }
        return 0;
    }
}
