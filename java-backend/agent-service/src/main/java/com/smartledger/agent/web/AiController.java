package com.smartledger.agent.web;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.smartledger.agent.security.GatewayUserFilter;
import com.smartledger.agent.service.AgentChatService;
import com.smartledger.agent.service.AgentStorageService;
import com.smartledger.agent.web.dto.AgentChatRequest;
import com.smartledger.agent.web.dto.AgentStorageLoadRequest;
import com.smartledger.agent.web.dto.AgentStorageLoadResponse;
import com.smartledger.agent.web.dto.AgentStorageSaveRequest;
import com.smartledger.agent.web.dto.TestAgentRequest;
import com.smartledger.common.web.ApiErrorAdvice.ApiException;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RestController;
import org.springframework.web.servlet.mvc.method.annotation.SseEmitter;

import java.io.IOException;
import java.util.LinkedHashMap;
import java.util.Map;
import java.util.concurrent.atomic.AtomicBoolean;

@RestController
public class AiController {

    private final AgentChatService chatService;
    private final AgentStorageService storageService;
    private final ObjectMapper mapper = new ObjectMapper();

    public AiController(AgentChatService chatService, AgentStorageService storageService) {
        this.chatService = chatService;
        this.storageService = storageService;
    }

    @GetMapping("/api/v1/health")
    public Map<String, Boolean> health() {
        return Map.of("ok", true);
    }

    @PostMapping("/api/v1/ai/test")
    public Map<String, Boolean> test(@RequestBody TestAgentRequest req) {
        GatewayUserFilter.GatewayUser user = GatewayUserFilter.currentUser();
        if (user == null) {
            throw ApiException.unauthorized("unauthorized");
        }
        validateLlm(req.baseUrl(), req.model());
        chatService.testConnection(req.baseUrl(), req.apiKey(), req.model());
        return Map.of("ok", true);
    }

    @PostMapping("/api/v1/ai/chat")
    public Object chat(@RequestBody AgentChatRequest req) throws IOException {
        GatewayUserFilter.GatewayUser user = GatewayUserFilter.currentUser();
        validateChat(req);
        rejectLegacyOpenClaw(req);

        if (req.stream()) {
            SseEmitter emitter = new SseEmitter(600_000L);
            AtomicBoolean started = new AtomicBoolean(false);
            Thread.startVirtualThread(() -> {
                try {
                    chatService.chatStream(req, user, delta -> {
                        try {
                            if (delta != null && !delta.isBlank()) {
                                started.set(true);
                                emitter.send(SseEmitter.event().data(ssePayload(delta)));
                            }
                        } catch (IOException e) {
                            emitter.completeWithError(e);
                        }
                    });
                    if (!started.get()) {
                        emitter.completeWithError(new IllegalStateException("LLM returned empty streaming response"));
                        return;
                    }
                    emitter.send(SseEmitter.event().data("[DONE]"));
                    emitter.complete();
                } catch (Exception e) {
                    if (!started.get()) {
                        emitter.completeWithError(e);
                        return;
                    }
                    try {
                        Map<String, Object> err = Map.of("error", Map.of("message", e.getMessage()));
                        emitter.send(SseEmitter.event().data(mapper.writeValueAsString(err)));
                        emitter.send(SseEmitter.event().data("[DONE]"));
                        emitter.complete();
                    } catch (IOException ex) {
                        emitter.completeWithError(ex);
                    }
                }
            });
            return emitter;
        }

        String content = chatService.chatJson(req, user);
        Map<String, Object> choice = new LinkedHashMap<>();
        choice.put("index", 0);
        choice.put("message", Map.of("role", "assistant", "content", content));
        choice.put("finish_reason", "stop");
        return Map.of("choices", java.util.List.of(choice));
    }

    @PostMapping("/api/v1/ai/agent/load")
    public AgentStorageLoadResponse load(@RequestBody AgentStorageLoadRequest req) {
        GatewayUserFilter.currentUser();
        if (isBlank(req.agentPath()) || isBlank(req.chatHistoryPath())) {
            throw ApiException.badRequest("agentPath and chatHistoryPath required");
        }
        return storageService.load(req);
    }

    @PostMapping("/api/v1/ai/agent/save")
    public Map<String, Boolean> save(@RequestBody AgentStorageSaveRequest req) {
        GatewayUserFilter.currentUser();
        if (isBlank(req.agentPath()) || isBlank(req.chatHistoryPath())) {
            throw ApiException.badRequest("agentPath and chatHistoryPath required");
        }
        if ((req.messages() == null || req.messages().isEmpty())
                && (req.workspaceFiles() == null || req.workspaceFiles().isEmpty())) {
            throw ApiException.badRequest("nothing to save");
        }
        storageService.save(req);
        return Map.of("ok", true);
    }

    private static String ssePayload(String delta) throws IOException {
        ObjectMapper m = new ObjectMapper();
        return m.writeValueAsString(Map.of(
                "choices", java.util.List.of(Map.of(
                        "index", 0,
                        "delta", Map.of("content", delta)))));
    }

    private static void validateChat(AgentChatRequest req) {
        validateLlm(req.baseUrl(), req.model());
        if (req.messages() == null || req.messages().isEmpty()) {
            throw ApiException.badRequest("messages required");
        }
    }

    private static void validateLlm(String baseUrl, String model) {
        if (isBlank(baseUrl) || isBlank(model)) {
            throw ApiException.badRequest("baseUrl and model required");
        }
    }

    private static void rejectLegacyOpenClaw(AgentChatRequest req) {
        if (!isBlank(req.baseUrl())) {
            return;
        }
        throw ApiException.badRequest(
                "openclaw gateway is deprecated; configure provider baseUrl/apiKey/model in Settings → AI");
    }

    private static boolean isBlank(String s) {
        return s == null || s.isBlank();
    }
}
