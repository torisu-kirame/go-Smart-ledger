package com.smartledger.common.jwt;

import io.jsonwebtoken.Claims;
import io.jsonwebtoken.Jwts;
import io.jsonwebtoken.security.Keys;
import org.springframework.stereotype.Service;

import javax.crypto.SecretKey;
import java.nio.charset.StandardCharsets;
import java.time.Instant;
import java.util.Date;

/**
 * JWT 签发与校验：双令牌模型（短效 access + 长效 refresh）。
 * <p>
 * Access 走 Authorization 头；Refresh 经 HttpOnly Cookie 下发，密钥分离。
 */
@Service
public class JwtService {

    /** 自定义 claim，网关解析后写入 X-User-Id / X-Username */
    public static final String CLAIM_UID = "uid";
    public static final String CLAIM_USR = "usr";
    public static final String CLAIM_TYP = "typ";
    public static final String TYPE_ACCESS = "access";
    public static final String TYPE_REFRESH = "refresh";

    public static final String REFRESH_COOKIE = "sl_refresh_token";

    private final JwtProperties props;
    private final SecretKey accessKey;
    private final SecretKey refreshKey;

    public JwtService(JwtProperties props) {
        this.props = props;
        this.accessKey = Keys.hmacShaKeyFor(props.accessSecret().getBytes(StandardCharsets.UTF_8));
        this.refreshKey = Keys.hmacShaKeyFor(props.refreshSecret().getBytes(StandardCharsets.UTF_8));
    }

    /** 同时签发 access（响应体）与 refresh（Cookie） */
    public TokenPair issue(String userId, String username) {
        Instant now = Instant.now();
        Instant accessExp = now.plusSeconds(props.accessExpireSeconds());
        Instant refreshExp = now.plusSeconds(props.refreshExpireSeconds());

        String access = Jwts.builder()
                .claim(CLAIM_UID, userId)
                .claim(CLAIM_USR, username)
                .claim(CLAIM_TYP, TYPE_ACCESS)
                .subject(userId)
                .issuedAt(Date.from(now))
                .expiration(Date.from(accessExp))
                .signWith(accessKey)
                .compact();

        String refresh = Jwts.builder()
                .claim(CLAIM_UID, userId)
                .claim(CLAIM_USR, username)
                .claim(CLAIM_TYP, TYPE_REFRESH)
                .subject(userId)
                .issuedAt(Date.from(now))
                .expiration(Date.from(refreshExp))
                .signWith(refreshKey)
                .compact();

        return new TokenPair(access, refresh, props.accessExpireSeconds());
    }

    public TokenPair refreshAccess(String refreshToken) {
        JwtUser user = parse(refreshKey, refreshToken, TYPE_REFRESH);
        return issue(user.userId(), user.username());
    }

    public JwtUser parseAccess(String token) {
        return parse(accessKey, token, TYPE_ACCESS);
    }

    public JwtUser parseRefresh(String token) {
        return parse(refreshKey, token, TYPE_REFRESH);
    }

    /** 校验签名与 typ，防止 refresh 被当作 access 使用 */
    private JwtUser parse(SecretKey key, String token, String expectedType) {
        Claims claims = Jwts.parser()
                .verifyWith(key)
                .build()
                .parseSignedClaims(token)
                .getPayload();
        String typ = claims.get(CLAIM_TYP, String.class);
        if (!expectedType.equals(typ)) {
            throw new IllegalArgumentException("invalid token type");
        }
        return new JwtUser(
                claims.get(CLAIM_UID, String.class),
                claims.get(CLAIM_USR, String.class)
        );
    }

    public record TokenPair(String accessToken, String refreshToken, long expiresIn) {}

    public record JwtUser(String userId, String username) {}
}
