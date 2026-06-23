package com.smartledger.agent.web.dto;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

import java.util.List;
import java.util.Map;

@JsonIgnoreProperties(ignoreUnknown = true)
public record AgentStorageLoadResponse(
        List<ChatMessage> messages,
        Map<String, String> workspaceFiles,
        String updatedAt
) {
}
