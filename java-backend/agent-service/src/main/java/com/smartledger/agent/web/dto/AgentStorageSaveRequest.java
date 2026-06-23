package com.smartledger.agent.web.dto;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

import java.util.List;
import java.util.Map;

@JsonIgnoreProperties(ignoreUnknown = true)
public record AgentStorageSaveRequest(
        String agentPath,
        String chatHistoryPath,
        List<ChatMessage> messages,
        Map<String, String> workspaceFiles
) {
}
