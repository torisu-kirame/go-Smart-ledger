package com.smartledger.agent.service;

import com.smartledger.agent.config.AgentProperties;
import com.smartledger.agent.llm.LlmProvider;
import com.smartledger.agent.security.GatewayUserFilter.GatewayUser;
import com.smartledger.agent.tools.LedgerClient;
import com.smartledger.agent.tools.LedgerTools;
import com.smartledger.agent.web.dto.AgentChatRequest;
import com.smartledger.agent.web.dto.ChatMessage;
import com.smartledger.agent.workspace.WorkspaceLoader;
import com.smartledger.common.web.ApiErrorAdvice.ApiException;
import dev.langchain4j.data.message.AiMessage;
import dev.langchain4j.data.message.SystemMessage;
import dev.langchain4j.data.message.UserMessage;
import dev.langchain4j.model.StreamingResponseHandler;
import dev.langchain4j.model.chat.ChatLanguageModel;
import dev.langchain4j.model.chat.StreamingChatLanguageModel;
import dev.langchain4j.model.output.Response;
import dev.langchain4j.service.AiServices;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.atomic.AtomicReference;
import java.util.function.Consumer;

/** LangChain4j ReAct Agent：Prompt 编排 + Tool 工作区 + 对话流程 */
@Service
public class AgentChatService {

    interface LedgerAgent {
        String chat(String userMessage);
    }

    private final AgentProperties props;
    private final LedgerClient ledgerClient;
    private final WorkspaceLoader workspaceLoader;

    public AgentChatService(AgentProperties props, LedgerClient ledgerClient, WorkspaceLoader workspaceLoader) {
        this.props = props;
        this.ledgerClient = ledgerClient;
        this.workspaceLoader = workspaceLoader;
    }

    public void testConnection(String baseUrl, String apiKey, String model) {
        ChatLanguageModel modelBean = LlmProvider.chatModel(baseUrl, apiKey, model);
        modelBean.generate(UserMessage.from("ping"));
    }

    public String chatJson(AgentChatRequest req, GatewayUser user) {
        if (req.useTools()) {
            return runReActAgent(req, user);
        }
        ChatLanguageModel llm = LlmProvider.chatModel(req.baseUrl(), req.apiKey(), req.model());
        Response<AiMessage> response = llm.generate(toLcMessages(req.messages(), workspaceLoader.defaultSystemPrompt()));
        return response.content().text();
    }

    public void chatStream(AgentChatRequest req, GatewayUser user, Consumer<String> onDelta) {
        if (req.useTools()) {
            String content = runReActAgent(req, user);
            chunkEmit(content, onDelta);
            return;
        }
        StreamingChatLanguageModel llm = LlmProvider.streamingModel(req.baseUrl(), req.apiKey(), req.model());
        CountDownLatch latch = new CountDownLatch(1);
        AtomicReference<Throwable> err = new AtomicReference<>();
        llm.generate(toLcMessages(req.messages(), workspaceLoader.defaultSystemPrompt()), new StreamingResponseHandler<>() {
            @Override
            public void onNext(String token) {
                if (token != null && !token.isBlank()) {
                    onDelta.accept(token);
                }
            }

            @Override
            public void onComplete(Response<AiMessage> response) {
                latch.countDown();
            }

            @Override
            public void onError(Throwable error) {
                err.set(error);
                latch.countDown();
            }
        });
        try {
            latch.await();
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            throw ApiException.badRequest("stream interrupted");
        }
        if (err.get() != null) {
            throw ApiException.badRequest(err.get().getMessage());
        }
    }

    private String runReActAgent(AgentChatRequest req, GatewayUser user) {
        String input = PromptBuilder.buildAgentInput(req.messages());
        if (input.isBlank()) {
            throw ApiException.badRequest("no user message for agent");
        }
        String systemPrefix = PromptBuilder.buildSystemPrefix(
                req.messages(), req.boundLedgerId(), workspaceLoader.defaultSystemPrompt());
        ChatLanguageModel llm = LlmProvider.chatModel(req.baseUrl(), req.apiKey(), req.model());
        LedgerTools tools = new LedgerTools(
                ledgerClient, user.userId(), user.authorizationHeader(), req.boundLedgerId());
        LedgerAgent agent = AiServices.builder(LedgerAgent.class)
                .chatLanguageModel(llm)
                .tools(tools)
                .systemMessageProvider(memoryId -> systemPrefix)
                .build();
        String content = agent.chat(input);
        if (content == null || content.isBlank()) {
            throw ApiException.badRequest("agent returned empty response");
        }
        return content.trim();
    }

    private static List<dev.langchain4j.data.message.ChatMessage> toLcMessages(
            List<ChatMessage> messages, String defaultSystemPrompt) {
        List<dev.langchain4j.data.message.ChatMessage> out = new ArrayList<>();
        boolean hasSystem = false;
        for (ChatMessage m : messages) {
            String role = m.role() == null ? "" : m.role().toLowerCase().trim();
            String content = m.content() == null ? "" : m.content().trim();
            if (content.isEmpty()) {
                continue;
            }
            switch (role) {
                case "system" -> {
                    hasSystem = true;
                    out.add(SystemMessage.from(content));
                }
                case "assistant" -> out.add(AiMessage.from(content));
                default -> out.add(UserMessage.from(content));
            }
        }
        if (!hasSystem && defaultSystemPrompt != null && !defaultSystemPrompt.isBlank()) {
            out.addFirst(SystemMessage.from(defaultSystemPrompt));
        }
        return out;
    }

    private static void chunkEmit(String text, Consumer<String> onDelta) {
        int size = 64;
        for (int i = 0; i < text.length(); i += size) {
            onDelta.accept(text.substring(i, Math.min(i + size, text.length())));
        }
    }
}
