package com.smartledger.agent.llm;

import com.smartledger.common.web.ApiErrorAdvice.ApiException;
import dev.langchain4j.model.chat.ChatLanguageModel;
import dev.langchain4j.model.chat.StreamingChatLanguageModel;
import dev.langchain4j.model.ollama.OllamaChatModel;
import dev.langchain4j.model.ollama.OllamaStreamingChatModel;
import dev.langchain4j.model.openai.OpenAiChatModel;
import dev.langchain4j.model.openai.OpenAiStreamingChatModel;

import java.net.InetAddress;
import java.net.URI;
import java.time.Duration;
import java.util.Set;

/** 云端 OpenAI 兼容 API 与本地 Ollama 双模式切换 */
public final class LlmProvider {

    private static final Set<String> LOCAL_HOSTS = Set.of(
            "127.0.0.1", "localhost", "::1", "ollama", "host.docker.internal");

    private LlmProvider() {
    }

    public static ChatLanguageModel chatModel(String baseUrl, String apiKey, String model) {
        if (isOllamaEndpoint(baseUrl)) {
            return OllamaChatModel.builder()
                    .baseUrl(stripOpenAiSuffix(baseUrl))
                    .modelName(model.trim())
                    .timeout(Duration.ofSeconds(120))
                    .build();
        }
        String norm = normalizeOpenAiBaseUrl(baseUrl);
        var builder = OpenAiChatModel.builder()
                .baseUrl(norm)
                .modelName(model.trim())
                .timeout(Duration.ofSeconds(120));
        if (apiKey != null && !apiKey.isBlank()) {
            builder.apiKey(apiKey.trim());
        }
        return builder.build();
    }

    public static StreamingChatLanguageModel streamingModel(String baseUrl, String apiKey, String model) {
        if (isOllamaEndpoint(baseUrl)) {
            return OllamaStreamingChatModel.builder()
                    .baseUrl(stripOpenAiSuffix(baseUrl))
                    .modelName(model.trim())
                    .timeout(Duration.ofSeconds(120))
                    .build();
        }
        String norm = normalizeOpenAiBaseUrl(baseUrl);
        var builder = OpenAiStreamingChatModel.builder()
                .baseUrl(norm)
                .modelName(model.trim())
                .timeout(Duration.ofSeconds(120));
        if (apiKey != null && !apiKey.isBlank()) {
            builder.apiKey(apiKey.trim());
        }
        return builder.build();
    }

    static boolean isOllamaEndpoint(String baseUrl) {
        if (baseUrl == null || baseUrl.isBlank()) {
            return false;
        }
        try {
            URI uri = URI.create(baseUrl.trim());
            String host = uri.getHost();
            if (host == null) {
                return false;
            }
            host = host.toLowerCase();
            if (LOCAL_HOSTS.contains(host)) {
                return true;
            }
            int port = uri.getPort();
            return port == 11434;
        } catch (Exception e) {
            return false;
        }
    }

    static String normalizeOpenAiBaseUrl(String raw) {
        String norm = ensureOpenAiV1Suffix(raw.trim());
        try {
            URI uri = URI.create(norm);
            String host = uri.getHost() == null ? "" : uri.getHost().toLowerCase();
            if ("https".equalsIgnoreCase(uri.getScheme())) {
                if (isPrivateHost(host)) {
                    throw ApiException.badRequest("ai base url not allowed");
                }
            } else if ("http".equalsIgnoreCase(uri.getScheme())) {
                if (!LOCAL_HOSTS.contains(host)) {
                    throw ApiException.badRequest("ai base url not allowed");
                }
            } else {
                throw ApiException.badRequest("ai base url not allowed");
            }
            return norm;
        } catch (ApiException e) {
            throw e;
        } catch (Exception e) {
            throw ApiException.badRequest("ai base url not allowed");
        }
    }

    private static boolean isPrivateHost(String host) {
        if (host.isBlank() || LOCAL_HOSTS.contains(host)) {
            return false;
        }
        try {
            InetAddress addr = InetAddress.getByName(host);
            return addr.isLoopbackAddress() || addr.isSiteLocalAddress()
                    || addr.isLinkLocalAddress() || addr.isMulticastAddress();
        } catch (Exception e) {
            return false;
        }
    }

    private static String ensureOpenAiV1Suffix(String raw) {
        raw = raw.replaceAll("/+$", "");
        if (raw.endsWith("/v1")) {
            return raw;
        }
        return raw + "/v1";
    }

    private static String stripOpenAiSuffix(String raw) {
        raw = raw.trim().replaceAll("/+$", "");
        if (raw.endsWith("/v1")) {
            return raw.substring(0, raw.length() - 3);
        }
        return raw;
    }
}
