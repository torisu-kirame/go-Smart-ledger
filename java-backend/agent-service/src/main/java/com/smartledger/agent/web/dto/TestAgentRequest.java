package com.smartledger.agent.web.dto;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;

@JsonIgnoreProperties(ignoreUnknown = true)
public record TestAgentRequest(String baseUrl, String apiKey, String model) {
}
