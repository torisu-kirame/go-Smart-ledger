package com.smartledger.ledger.web;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestHeader;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.client.RestTemplate;
import com.smartledger.common.web.ApiErrorAdvice.ApiException;

import java.util.Collections;
import java.util.List;
import java.util.Map;

/**
 * 账本与链相关 API（开发阶段桩实现）。
 * <p>
 * 当前支持健康检查与空列表查询；完整 CRUD 与链上写入待后续迭代。
 */
@RestController
public class LedgerController {
    private final RestTemplate rest = new RestTemplate();
    private final String miniledgerBase;

    public LedgerController(@Value("${smartledger.miniledger.base-url:http://miniledger:24441}") String baseUrl) {
        this.miniledgerBase = baseUrl.replaceAll("/$", "");
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
                "queueFailed", 0
        );
    }

    /**
     * 返回空数组，保证登录后列表页可正常渲染。
     * 用户身份由网关注入的 X-User-Id 提供。
     */
    @GetMapping("/api/v1/ledgers")
    public List<?> listLedgers(@RequestHeader(value = "X-User-Id", required = false) String userId) {
        if (userId == null || userId.isBlank()) {
            throw ApiException.unauthorized("unauthorized");
        }
        return Collections.emptyList();
    }

    @GetMapping("/api/v1/chain/status")
    public Map<String, Object> chainStatus() {
        boolean online = pingMiniLedger();
        return Map.of(
                "backend", "miniledger",
                "online", online,
                "impl", "java-stub"
        );
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
