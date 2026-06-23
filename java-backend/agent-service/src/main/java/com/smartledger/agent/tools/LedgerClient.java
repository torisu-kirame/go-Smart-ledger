package com.smartledger.agent.tools;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.smartledger.agent.config.AgentProperties;
import org.springframework.http.HttpEntity;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpMethod;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Component;
import org.springframework.web.client.RestTemplate;

import java.util.List;
import java.util.Map;

/**
 * 调用 ledger-api 的 HTTP 客户端，转发 JWT 与 X-User-Id 实现成员级账本隔离。
 */
@Component
public class LedgerClient {

    private final RestTemplate rest = new RestTemplate();
    private final ObjectMapper mapper = new ObjectMapper();
    private final String baseUrl;

    public LedgerClient(AgentProperties props) {
        this.baseUrl = props.ledgerApiUrl().replaceAll("/+$", "");
    }

    public List<Map<String, Object>> listLedgers(String userId, String authorization) {
        return exchangeList("/api/v1/ledgers", userId, authorization);
    }

    public Map<String, Object> getLedger(String ledgerId, String userId, String authorization) {
        return exchangeMap("/api/v1/ledgers/" + ledgerId, userId, authorization);
    }

    public Map<String, Object> exportRag(String ledgerId, String userId, String authorization) {
        return exchangeMap("/api/v1/ledgers/" + ledgerId + "/rag-export", userId, authorization);
    }

    public Map<String, Object> verifyLedger(String ledgerId, String userId, String authorization) {
        return exchangeMap("/api/v1/ledgers/" + ledgerId + "/verify", userId, authorization);
    }

    private List<Map<String, Object>> exchangeList(String path, String userId, String authorization) {
        ResponseEntity<String> resp = rest.exchange(
                baseUrl + path,
                HttpMethod.GET,
                entity(userId, authorization),
                String.class);
        try {
            return mapper.readValue(resp.getBody(), new TypeReference<>() {
            });
        } catch (Exception e) {
            throw new IllegalStateException("ledger list parse error: " + e.getMessage());
        }
    }

    private Map<String, Object> exchangeMap(String path, String userId, String authorization) {
        ResponseEntity<String> resp = rest.exchange(
                baseUrl + path,
                HttpMethod.GET,
                entity(userId, authorization),
                String.class);
        try {
            return mapper.readValue(resp.getBody(), new TypeReference<>() {
            });
        } catch (Exception e) {
            throw new IllegalStateException("ledger response parse error: " + e.getMessage());
        }
    }

    private static HttpEntity<Void> entity(String userId, String authorization) {
        HttpHeaders headers = new HttpHeaders();
        if (authorization != null && !authorization.isBlank()) {
            headers.set(HttpHeaders.AUTHORIZATION, authorization);
        }
        if (userId != null && !userId.isBlank()) {
            headers.set("X-User-Id", userId);
        }
        return new HttpEntity<>(headers);
    }
}
