package com.smartledger.agent.security;

import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.security.authentication.UsernamePasswordAuthenticationToken;
import org.springframework.security.core.authority.SimpleGrantedAuthority;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

import java.io.IOException;
import java.util.List;

/**
 * 从网关注入的身份头构建 Spring Security 上下文。
 * <p>
 * JWT 已在 gateway 验签；本服务只信任内网 {@code X-User-Id}，用于 Tool 调用的成员级账本隔离。
 */
@Component
public class GatewayUserFilter extends OncePerRequestFilter {

    public static final String HDR_USER_ID = "X-User-Id";
    public static final String HDR_USERNAME = "X-Username";
    public static final String HDR_AUTHORIZATION = "Authorization";

    @Override
    protected void doFilterInternal(HttpServletRequest request, HttpServletResponse response, FilterChain chain)
            throws ServletException, IOException {
        String path = request.getRequestURI();
        if ("/api/v1/health".equals(path)) {
            chain.doFilter(request, response);
            return;
        }
        String userId = header(request, HDR_USER_ID);
        if (userId == null || userId.isBlank()) {
            writeJsonError(response, 401, "unauthorized");
            return;
        }
        String username = header(request, HDR_USERNAME);
        var auth = new UsernamePasswordAuthenticationToken(
                userId,
                null,
                List.of(new SimpleGrantedAuthority("ROLE_MEMBER")));
        auth.setDetails(new GatewayUser(userId, username, header(request, HDR_AUTHORIZATION)));
        SecurityContextHolder.getContext().setAuthentication(auth);
        chain.doFilter(request, response);
    }

    private static String header(HttpServletRequest request, String name) {
        String v = request.getHeader(name);
        return v == null ? null : v.trim();
    }

    private static void writeJsonError(HttpServletResponse response, int code, String msg) throws IOException {
        response.setStatus(code);
        response.setContentType("application/json;charset=UTF-8");
        response.getWriter().write("{\"code\":" + code + ",\"msg\":\"" + msg + "\"}");
    }

    public record GatewayUser(String userId, String username, String authorizationHeader) {
    }

    public static GatewayUser currentUser() {
        var auth = SecurityContextHolder.getContext().getAuthentication();
        if (auth == null || !(auth.getDetails() instanceof GatewayUser user)) {
            throw new IllegalStateException("unauthorized");
        }
        return user;
    }
}
