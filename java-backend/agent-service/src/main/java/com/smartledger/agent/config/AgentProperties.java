package com.smartledger.agent.config;

import org.springframework.boot.context.properties.ConfigurationProperties;

@ConfigurationProperties(prefix = "smartledger.agent")
public record AgentProperties(
        String configPath,
        String ledgerApiUrl,
        int maxAgentIterations
) {
    public AgentProperties {
        if (configPath == null || configPath.isBlank()) {
            configPath = "/data/agent/config";
        }
        if (ledgerApiUrl == null || ledgerApiUrl.isBlank()) {
            ledgerApiUrl = "http://ledger-api:28888";
        }
        if (maxAgentIterations <= 0) {
            maxAgentIterations = 6;
        }
    }
}
