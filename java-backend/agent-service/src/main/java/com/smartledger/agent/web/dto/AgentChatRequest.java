package com.smartledger.agent.web.dto;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

import java.util.ArrayList;
import java.util.List;

@JsonIgnoreProperties(ignoreUnknown = true)
public record AgentChatRequest(
        String baseUrl,
        String apiKey,
        String model,
        List<ChatMessage> messages,
        boolean stream,
        boolean useTools,
        String boundLedgerId
) {
    public AgentChatRequest {
        if (messages == null) {
            messages = new ArrayList<>();
        }
    }
}
