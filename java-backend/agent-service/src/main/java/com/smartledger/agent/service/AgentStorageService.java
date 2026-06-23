package com.smartledger.agent.service;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.smartledger.agent.config.AgentProperties;
import com.smartledger.agent.web.dto.AgentStorageLoadRequest;
import com.smartledger.agent.web.dto.AgentStorageLoadResponse;
import com.smartledger.agent.web.dto.AgentStorageSaveRequest;
import com.smartledger.agent.web.dto.ChatMessage;
import com.smartledger.agent.workspace.WorkspaceLoader;
import com.smartledger.common.web.ApiErrorAdvice.ApiException;
import org.springframework.stereotype.Service;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.time.Instant;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Set;

/** Agent 会话与工作区 Markdown 持久化 */
@Service
public class AgentStorageService {

    private static final String CHAT_FILE = "messages.json";
    private static final Set<String> ALLOWED_WORKSPACE = Set.of(
            "AGENTS.md", "BOOTSTRAP.md", "HEARTBEAT.md", "IDENTITY.md",
            "SOUL.md", "TOOLS.md", "USER.md");

    private final Path configRoot;
    private final WorkspaceLoader workspaceLoader;
    private final ObjectMapper mapper = new ObjectMapper();

    public AgentStorageService(AgentProperties props, WorkspaceLoader workspaceLoader) {
        this.configRoot = Paths.get(props.configPath()).toAbsolutePath().normalize();
        this.workspaceLoader = workspaceLoader;
    }

    public AgentStorageLoadResponse load(AgentStorageLoadRequest req) {
        Path chatDir = resolveSafe(req.chatHistoryPath());
        List<ChatMessage> messages = List.of();
        String updatedAt = "";
        Path chatPath = chatDir.resolve(CHAT_FILE);
        if (Files.isRegularFile(chatPath)) {
            try {
                @SuppressWarnings("unchecked")
                Map<String, Object> payload = mapper.readValue(chatPath.toFile(), Map.class);
                @SuppressWarnings("unchecked")
                List<Map<String, String>> raw = (List<Map<String, String>>) payload.get("messages");
                if (raw != null) {
                    messages = raw.stream()
                            .map(m -> new ChatMessage(m.get("role"), m.get("content")))
                            .toList();
                }
                updatedAt = stringVal(payload.get("updatedAt"));
            } catch (IOException e) {
                throw ApiException.badRequest("agent storage error: read chat");
            }
        }
        Map<String, String> ws = new LinkedHashMap<>();
        if (req.loadWorkspace()) {
            Path wsDir = workspaceDir(req.agentPath());
            for (String name : ALLOWED_WORKSPACE) {
                Path fp = wsDir.resolve(name);
                if (Files.isRegularFile(fp)) {
                    try {
                        ws.put(name, Files.readString(fp));
                    } catch (IOException ignored) {
                    }
                }
            }
            if (ws.isEmpty()) {
                ws.putAll(defaultWorkspaceFiles());
            }
        }
        return new AgentStorageLoadResponse(messages, ws, updatedAt);
    }

    public void save(AgentStorageSaveRequest req) {
        if (req.messages() != null && !req.messages().isEmpty()) {
            Path chatDir = resolveSafe(req.chatHistoryPath());
            try {
                Files.createDirectories(chatDir);
                Map<String, Object> payload = Map.of(
                        "version", 1,
                        "updatedAt", Instant.now().toString(),
                        "messages", req.messages());
                mapper.writerWithDefaultPrettyPrinter().writeValue(chatDir.resolve(CHAT_FILE).toFile(), payload);
            } catch (IOException e) {
                throw ApiException.badRequest("agent storage error: write chat");
            }
        }
        if (req.workspaceFiles() != null && !req.workspaceFiles().isEmpty()) {
            Path wsDir = workspaceDir(req.agentPath());
            try {
                Files.createDirectories(wsDir);
                for (Map.Entry<String, String> e : req.workspaceFiles().entrySet()) {
                    if (ALLOWED_WORKSPACE.contains(e.getKey())) {
                        Files.writeString(wsDir.resolve(e.getKey()), e.getValue() == null ? "" : e.getValue());
                    }
                }
            } catch (IOException ex) {
                throw ApiException.badRequest("agent storage error: write workspace");
            }
        }
    }

    public Map<String, String> defaultWorkspaceFiles() {
        Map<String, String> out = new LinkedHashMap<>();
        String prompt = workspaceLoader.defaultSystemPrompt();
        if (!prompt.isBlank()) {
            out.put("AGENTS.md", prompt.split("---")[0].trim());
            if (prompt.contains("---")) {
                out.put("TOOLS.md", prompt.substring(prompt.indexOf("---") + 3).trim());
            }
        }
        return out;
    }

    private Path workspaceDir(String agentPath) {
        String normalized = agentPath.replace('\\', '/').replaceAll("^/+|/+$", "");
        if ("agents/main".equals(normalized)) {
            Path legacy = configRoot.resolve("workspace/workspace-smart-ledger");
            if (Files.isDirectory(legacy)) {
                return legacy;
            }
        }
        return configRoot.resolve(normalized).resolve("workspace");
    }

    private Path resolveSafe(String relPath) {
        if (relPath == null || relPath.isBlank()) {
            throw ApiException.badRequest("agent path not allowed");
        }
        relPath = relPath.replace('\\', '/').trim();
        if (relPath.startsWith("/") || relPath.contains("..")) {
            throw ApiException.badRequest("agent path not allowed: " + relPath);
        }
        Path full = configRoot.resolve(relPath).normalize();
        if (!full.startsWith(configRoot)) {
            throw ApiException.badRequest("agent path not allowed: " + relPath);
        }
        return full;
    }

    private static String stringVal(Object v) {
        return v == null ? "" : v.toString();
    }
}
