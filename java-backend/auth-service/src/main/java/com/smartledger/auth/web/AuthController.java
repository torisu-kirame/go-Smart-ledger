package com.smartledger.auth.web;

import com.smartledger.auth.service.UserService;
import com.smartledger.auth.service.UserService.User;
import com.smartledger.common.captcha.CaptchaService;
import com.smartledger.common.jwt.JwtProperties;
import com.smartledger.common.jwt.JwtService;
import com.smartledger.common.jwt.JwtService.TokenPair;
import com.smartledger.common.web.ApiErrorAdvice.ApiException;
import jakarta.servlet.http.Cookie;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.http.ResponseCookie;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.Map;

/**
 * 认证 API：登录、注册、令牌刷新。
 * <p>
 * Access Token 走 Authorization 头；Refresh Token 写入 HttpOnly Cookie，
 * 前端 {@code credentials: 'include'} 即可自动携带。
 */
@RestController
@RequestMapping("/api/v1/auth")
public class AuthController {
    private final CaptchaService captchaService;
    private final UserService userService;
    private final JwtService jwtService;
    private final JwtProperties jwtProperties;

    public AuthController(CaptchaService captchaService, UserService userService,
                          JwtService jwtService, JwtProperties jwtProperties) {
        this.captchaService = captchaService;
        this.userService = userService;
        this.jwtService = jwtService;
        this.jwtProperties = jwtProperties;
    }

    @GetMapping("/health")
    public Map<String, String> health() {
        return Map.of("status", "ok", "backend", "java");
    }

    @GetMapping("/captcha")
    public Map<String, String> captcha() {
        var c = captchaService.generate();
        return Map.of("captchaId", c.captchaId(), "image", c.image());
    }

    @PostMapping("/login")
    public LoginResp login(@RequestBody LoginReq req, HttpServletResponse response) {
        validateCaptcha(req.captchaId(), req.captchaCode());
        if (req.username() == null || req.username().isBlank() || req.password() == null || req.password().isBlank()) {
            throw ApiException.badRequest("username and password required");
        }
        try {
            User user = userService.authenticate(req.username(), req.password());
            return issueTokens(user, response);
        } catch (IllegalArgumentException e) {
            // 凭证错误统一 401，避免通过响应差异枚举用户名
            throw ApiException.unauthorized("invalid username or password");
        }
    }

    @PostMapping("/register")
    public LoginResp register(@RequestBody RegisterReq req, HttpServletResponse response) {
        validateCaptcha(req.captchaId(), req.captchaCode());
        if (req.username() == null || req.username().isBlank() || req.password() == null || req.password().length() < 6) {
            throw ApiException.badRequest("invalid register payload");
        }
        try {
            User user = userService.register(req.username(), req.password());
            return issueTokens(user, response);
        } catch (IllegalStateException e) {
            throw ApiException.badRequest(e.getMessage());
        }
    }

    /** 刷新时轮换 refresh 令牌；未做黑名单，单机场景可接受 */
    @PostMapping("/refresh")
    public LoginResp refresh(HttpServletRequest request, HttpServletResponse response) {
        String token = refreshFromCookie(request);
        if (token == null || token.isBlank()) {
            throw ApiException.unauthorized("missing refresh token");
        }
        try {
            var user = jwtService.parseRefresh(token);
            TokenPair pair = jwtService.issue(user.userId(), user.username());
            setRefreshCookie(response, pair.refreshToken());
            return new LoginResp(
                    pair.accessToken(),
                    pair.expiresIn(),
                    "Bearer",
                    new UserInfo(user.userId(), user.username(), avatarPath(user.userId()))
            );
        } catch (Exception e) {
            throw ApiException.unauthorized("invalid or expired refresh token");
        }
    }

    @PostMapping("/logout")
    public Map<String, Boolean> logout(HttpServletResponse response) {
        clearRefreshCookie(response);
        return Map.of("ok", true);
    }

    private LoginResp issueTokens(User user, HttpServletResponse response) {
        TokenPair pair = jwtService.issue(user.id(), user.username());
        setRefreshCookie(response, pair.refreshToken());
        return new LoginResp(
                pair.accessToken(),
                pair.expiresIn(),
                "Bearer",
                new UserInfo(user.id(), user.username(), avatarPath(user.id()))
        );
    }

    private static String avatarPath(String userId) {
        return "/api/v1/users/" + userId + "/avatar";
    }

    private void validateCaptcha(String captchaId, String captchaCode) {
        if (!captchaService.verify(captchaId, captchaCode, true)) {
            throw ApiException.badRequest("invalid captcha");
        }
    }

    private String refreshFromCookie(HttpServletRequest request) {
        Cookie[] cookies = request.getCookies();
        if (cookies == null) {
            return null;
        }
        for (Cookie c : cookies) {
            if (JwtService.REFRESH_COOKIE.equals(c.getName())) {
                return c.getValue();
            }
        }
        return null;
    }

    /** path 限定在 /api/v1/auth，refresh 请求不会把 Cookie 发送到业务接口 */
    private void setRefreshCookie(HttpServletResponse response, String token) {
        ResponseCookie cookie = ResponseCookie.from(JwtService.REFRESH_COOKIE, token)
                .httpOnly(true)
                .secure(jwtProperties.cookieSecure())
                .path("/api/v1/auth")
                .maxAge(jwtProperties.refreshExpireSeconds())
                .sameSite("Lax")
                .build();
        response.addHeader("Set-Cookie", cookie.toString());
    }

    private void clearRefreshCookie(HttpServletResponse response) {
        ResponseCookie cookie = ResponseCookie.from(JwtService.REFRESH_COOKIE, "")
                .httpOnly(true)
                .secure(jwtProperties.cookieSecure())
                .path("/api/v1/auth")
                .maxAge(0)
                .sameSite("Lax")
                .build();
        response.addHeader("Set-Cookie", cookie.toString());
    }

    public record LoginReq(String username, String password, String captchaId, String captchaCode) {}

    public record RegisterReq(String username, String password, String captchaId, String captchaCode) {}

    public record LoginResp(String accessToken, long expiresIn, String tokenType, UserInfo user) {}

    public record UserInfo(String id, String username, String avatarUrl) {}
}
