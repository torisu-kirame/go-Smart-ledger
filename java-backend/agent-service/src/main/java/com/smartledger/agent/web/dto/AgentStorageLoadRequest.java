package com.smartledger.agent.web.dto;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

@JsonIgnoreProperties(ignoreUnknown = true)
public record AgentStorageLoadRequest(String agentPath, String chatHistoryPath, boolean loadWorkspace) {
}
