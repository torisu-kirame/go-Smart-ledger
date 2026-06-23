package com.smartledger.agent.service;

import com.smartledger.agent.web.dto.ChatMessage;

import java.util.ArrayList;
import java.util.List;

final class PromptBuilder {

    private PromptBuilder() {
    }

    static String buildAgentInput(List<ChatMessage> messages) {
        int maxTurns = 10;
        List<String> turns = new ArrayList<>();
        for (ChatMessage m : messages) {
            String role = m.role() == null ? "" : m.role().toLowerCase().trim();
            String content = m.content() == null ? "" : m.content().trim();
            if ("system".equals(role) || content.isEmpty()) {
                continue;
            }
            String label = "assistant".equals(role) ? "Assistant" : "User";
            turns.add(label + ": " + content);
        }
        if (turns.isEmpty()) {
            return "";
        }
        if (turns.size() > maxTurns) {
            turns = turns.subList(turns.size() - maxTurns, turns.size());
        }
        String last = turns.getLast();
        if (!last.startsWith("User: ")) {
            return last;
        }
        if (turns.size() == 1) {
            return last.substring("User: ".length());
        }
        StringBuilder b = new StringBuilder("Conversation:\n");
        for (int i = 0; i < turns.size() - 1; i++) {
            b.append(turns.get(i)).append('\n');
        }
        b.append("\nCurrent question:\n").append(last.substring("User: ".length()));
        return b.toString();
    }

    static String buildSystemPrefix(List<ChatMessage> messages, String boundLedgerId, String defaultPrompt) {
        List<String> parts = new ArrayList<>();
        for (ChatMessage m : messages) {
            if (m.role() == null || !"system".equalsIgnoreCase(m.role().trim())) {
                continue;
            }
            String c = m.content() == null ? "" : m.content().trim();
            if (c.isEmpty()) {
                continue;
            }
            if (c.contains("账本上下文") || c.contains("RAG 导出")) {
                parts.add("用户已绑定账本；请用 search_ledger_rag / get_ledger_summary 等工具查询最新链上数据。");
                continue;
            }
            if (c.length() > 4000) {
                c = c.substring(0, 4000) + "…";
            }
            parts.add(c);
        }
        if (parts.isEmpty() && defaultPrompt != null && !defaultPrompt.isBlank()) {
            parts.add(defaultPrompt);
        }
        if (boundLedgerId != null && !boundLedgerId.isBlank()) {
            parts.add("当前绑定账本 ID：" + boundLedgerId.trim()
                    + "。调用需要 ledgerId 的工具时，若用户未指定其他账本，请使用该 ID。");
        }
        parts.add("你可以通过 list_ledgers、get_ledger_summary、search_ledger_rag、verify_ledger 工具查询账本；不要编造链上数据。");
        return String.join("\n\n", parts);
    }
}
