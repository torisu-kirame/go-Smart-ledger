package com.smartledger.ledger.web;

import com.smartledger.common.web.ApiErrorAdvice.ApiException;
import com.smartledger.ledger.service.LedgerStubStore;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestHeader;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.client.RestTemplate;

import java.util.List;
import java.util.Map;

/**
 * 账本与链相关 API（开发阶段桩实现，含 Agent Tool 所需 rag-export / verify）。
 */
@RestController
public class LedgerController {
    private final RestTemplate rest = new RestTemplate();
    private final String miniledgerBase;
    private final LedgerStubStore store;

    public LedgerController(
            @Value("${smartledger.miniledger.base-url:http://miniledger:24441}") String baseUrl,
            LedgerStubStore store) {
        this.miniledgerBase = baseUrl.replaceAll("/$", "");
        this.store = store;
    }

    @GetMapping("/api/v1/health")
    public Map<String, Object> health() {
        boolean online = pingMiniLedger();
        return Map.of(
                "status", "ok",
                "backend", "java",
                "miniLedgerOnline", online,
                "ipfsOnline", false,
                "queuePending", 0,
                "queueFailed", 0);
    }

    @GetMapping("/api/v1/ledgers")
    public List<Map<String, Object>> listLedgers(@RequestHeader(value = "X-User-Id", required = false) String userId) {
        requireUser(userId);
        return store.listForUser(userId);
    }

    @GetMapping("/api/v1/ledgers/{id}")
    public Map<String, Object> getLedger(
            @PathVariable("id") String id,
            @RequestHeader(value = "X-User-Id", required = false) String userId) {
        requireUser(userId);
        try {
            return store.getForUser(userId, id);
        } catch (IllegalArgumentException e) {
            throw ApiException.badRequest(e.getMessage());
        }
    }

    @GetMapping("/api/v1/ledgers/{id}/rag-export")
    public Map<String, Object> ragExport(
            @PathVariable("id") String id,
            @RequestHeader(value = "X-User-Id", required = false) String userId) {
        requireUser(userId);
        try {
            return store.ragExport(userId, id);
        } catch (IllegalArgumentException e) {
            throw ApiException.badRequest(e.getMessage());
        }
    }

    @GetMapping("/api/v1/ledgers/{id}/verify")
    public Map<String, Object> verifyLedger(
            @PathVariable("id") String id,
            @RequestHeader(value = "X-User-Id", required = false) String userId) {
        requireUser(userId);
        try {
            return store.verify(userId, id);
        } catch (IllegalArgumentException e) {
            throw ApiException.badRequest(e.getMessage());
        }
    }

    @GetMapping("/api/v1/chain/status")
    public Map<String, Object> chainStatus() {
        boolean online = pingMiniLedger();
        return Map.of(
                "backend", "miniledger",
                "online", online,
                "impl", "java-stub");
    }

    private static void requireUser(String userId) {
        if (userId == null || userId.isBlank()) {
            throw ApiException.unauthorized("unauthorized");
        }
    }

    private boolean pingMiniLedger() {
        try {
            Map<?, ?> body = rest.getForObject(miniledgerBase + "/status", Map.class);
            return body != null;
        } catch (Exception e) {
            return false;
        }
    }
}
