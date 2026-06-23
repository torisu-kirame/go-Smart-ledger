package com.smartledger.ledger.service;

import org.springframework.stereotype.Service;

import java.time.Instant;
import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

/**
 * 开发阶段内存账本，供 Agent Tool 联调；按 userId 隔离成员数据。
 */
@Service
public class LedgerStubStore {

    private final Map<String, List<Map<String, Object>>> userLedgers = new ConcurrentHashMap<>();

    public List<Map<String, Object>> listForUser(String userId) {
        return new ArrayList<>(userLedgers.computeIfAbsent(userId, this::seedLedgers));
    }

    public Map<String, Object> getForUser(String userId, String ledgerId) {
        return listForUser(userId).stream()
                .filter(l -> ledgerId.equals(l.get("id")))
                .findFirst()
                .orElseThrow(() -> new IllegalArgumentException("ledger not found"));
    }

    public Map<String, Object> ragExport(String userId, String ledgerId) {
        Map<String, Object> meta = getForUser(userId, ledgerId);
        List<Map<String, Object>> chunks = List.of(
                chunk(1, "ledger_meta", "账本「" + meta.get("name") + "」创建于 " + meta.get("createdAt")),
                chunk(2, "entry_added", "2026-05-01 报销交通费 128.50 元，科目：管理费用-差旅费"),
                chunk(3, "entry_added", "2026-05-03 收到客户货款 5000.00 元，银行存款增加"),
                chunk(4, "anchor", "Merkle 根已锚定 miniledger，anchorStatus=" + meta.get("anchorStatus")));
        Map<String, Object> out = new LinkedHashMap<>();
        out.put("ledgerId", ledgerId);
        out.put("ledgerName", meta.get("name"));
        out.put("exportedAt", Instant.now().toString());
        out.put("chunks", chunks);
        return out;
    }

    public Map<String, Object> verify(String userId, String ledgerId) {
        getForUser(userId, ledgerId);
        return Map.of("valid", true, "ledgerId", ledgerId);
    }

    private List<Map<String, Object>> seedLedgers(String userId) {
        String id = "ledger-demo-" + userId.substring(0, Math.min(8, userId.length()));
        Map<String, Object> ledger = new LinkedHashMap<>();
        ledger.put("id", id);
        ledger.put("name", "演示账本");
        ledger.put("type", "simple");
        ledger.put("latestSeq", 4L);
        ledger.put("latestRoot", "0xabc123demo");
        ledger.put("anchorStatus", "anchored");
        ledger.put("memberCount", 1);
        ledger.put("createdAt", "2026-05-01T00:00:00Z");
        ledger.put("updatedAt", Instant.now().toString());
        return new ArrayList<>(List.of(ledger));
    }

    private static Map<String, Object> chunk(long seq, String type, String text) {
        Map<String, Object> c = new LinkedHashMap<>();
        c.put("seq", seq);
        c.put("type", type);
        c.put("text", text);
        return c;
    }
}
