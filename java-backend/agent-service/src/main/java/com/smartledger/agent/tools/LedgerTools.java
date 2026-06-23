package com.smartledger.agent.tools;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import dev.langchain4j.agent.tool.Tool;

import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * LangChain4j ReAct Agent 账本工具集。
 * <p>
 * 每次 Tool 调用携带当前用户身份，ledger-api 按成员权限返回数据。
 */
public class LedgerTools {

    private final LedgerClient client;
    private final String userId;
    private final String authorization;
    private final String defaultLedgerId;
    private final ObjectMapper mapper = new ObjectMapper();

    public LedgerTools(LedgerClient client, String userId, String authorization, String defaultLedgerId) {
        this.client = client;
        this.userId = userId;
        this.authorization = authorization;
        this.defaultLedgerId = defaultLedgerId == null ? "" : defaultLedgerId.trim();
    }

    @Tool("List ledgers the current user can access. Input is optional JSON {} or empty.")
    public String listLedgers() {
        try {
            List<Map<String, Object>> rows = client.listLedgers(userId, authorization);
            return mapper.writeValueAsString(rows);
        } catch (Exception e) {
            return "error: " + e.getMessage();
        }
    }

    @Tool("Get metadata for one ledger. Input JSON {\"ledgerId\":\"...\"} or plain ledger id. Uses bound ledger when omitted.")
    public String getLedgerSummary(String ledgerId) {
        try {
            String id = resolveLedgerId(ledgerId);
            Map<String, Object> summary = client.getLedger(id, userId, authorization);
            return mapper.writeValueAsString(summary);
        } catch (Exception e) {
            return "error: " + e.getMessage();
        }
    }

    @Tool("Export ledger events as text chunks. Input JSON {\"ledgerId\":\"...\",\"limit\":40,\"query\":\"keyword\"}.")
    public String searchLedgerRag(String input) {
        try {
            Map<String, Object> args = parseArgs(input);
            String id = resolveLedgerId(stringVal(args.get("ledgerId"), input));
            int limit = intVal(args.get("limit"), 40);
            String query = stringVal(args.get("query"), "").toLowerCase();
            if (limit <= 0) {
                limit = 40;
            }
            if (limit > 120) {
                limit = 120;
            }
            Map<String, Object> export = client.exportRag(id, userId, authorization);
            @SuppressWarnings("unchecked")
            List<Map<String, Object>> chunks = (List<Map<String, Object>>) export.getOrDefault("chunks", List.of());
            List<Map<String, Object>> filtered = new ArrayList<>();
            for (Map<String, Object> c : chunks) {
                if (query.isEmpty()) {
                    filtered.add(c);
                    continue;
                }
                Object text = c.get("text");
                if (text != null && text.toString().toLowerCase().contains(query)) {
                    filtered.add(c);
                }
            }
            if (filtered.size() > limit) {
                filtered = filtered.subList(0, limit);
            }
            Map<String, Object> out = new LinkedHashMap<>();
            out.put("ledgerId", export.get("ledgerId"));
            out.put("ledgerName", export.get("ledgerName"));
            out.put("total", chunks.size());
            out.put("returned", filtered.size());
            out.put("chunks", filtered);
            String s = mapper.writeValueAsString(out);
            if (s.length() > 16000) {
                return s.substring(0, 16000) + "...(truncated)";
            }
            return s;
        } catch (Exception e) {
            return "error: " + e.getMessage();
        }
    }

    @Tool("Verify Merkle integrity. Input JSON {\"ledgerId\":\"...\"} or plain ledger id.")
    public String verifyLedger(String ledgerId) {
        try {
            String id = resolveLedgerId(ledgerId);
            client.getLedger(id, userId, authorization);
            Map<String, Object> out = client.verifyLedger(id, userId, authorization);
            return mapper.writeValueAsString(out);
        } catch (Exception e) {
            return "error: " + e.getMessage();
        }
    }

    private String resolveLedgerId(String input) {
        input = input == null ? "" : input.trim();
        if (input.startsWith("{")) {
            try {
                Map<String, Object> args = mapper.readValue(input, new TypeReference<>() {
                });
                String id = stringVal(args.get("ledgerId"), "");
                if (!id.isBlank()) {
                    return id;
                }
            } catch (Exception ignored) {
            }
        } else if (!input.isBlank()) {
            return input;
        }
        if (!defaultLedgerId.isBlank()) {
            return defaultLedgerId;
        }
        throw new IllegalArgumentException("ledgerId required");
    }

    private Map<String, Object> parseArgs(String input) {
        if (input == null || input.isBlank()) {
            return Map.of();
        }
        input = input.trim();
        if (!input.startsWith("{")) {
            return Map.of("ledgerId", input);
        }
        try {
            return mapper.readValue(input, new TypeReference<>() {
            });
        } catch (Exception e) {
            return Map.of();
        }
    }

    private static String stringVal(Object v, String fallback) {
        if (v == null) {
            return fallback == null ? "" : fallback;
        }
        String s = v.toString().trim();
        return s.isEmpty() && fallback != null ? fallback : s;
    }

    private static int intVal(Object v, int fallback) {
        if (v instanceof Number n) {
            return n.intValue();
        }
        if (v != null) {
            try {
                return Integer.parseInt(v.toString());
            } catch (NumberFormatException ignored) {
            }
        }
        return fallback;
    }
}
