package com.smartledger.agent.workspace;

import org.springframework.core.io.ClassPathResource;
import org.springframework.stereotype.Component;

import java.io.IOException;
import java.nio.charset.StandardCharsets;

@Component
public class WorkspaceLoader {

    public String defaultSystemPrompt() {
        StringBuilder sb = new StringBuilder();
        for (String name : new String[]{"workspace/AGENTS.md", "workspace/TOOLS.md"}) {
            try {
                String text = new ClassPathResource(name).getContentAsString(StandardCharsets.UTF_8);
                if (!text.isBlank()) {
                    if (!sb.isEmpty()) {
                        sb.append("\n\n---\n\n");
                    }
                    sb.append(text.trim());
                }
            } catch (IOException ignored) {
            }
        }
        return sb.toString().trim();
    }
}
