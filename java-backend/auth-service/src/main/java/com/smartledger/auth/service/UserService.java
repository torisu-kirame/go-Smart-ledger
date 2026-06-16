package com.smartledger.auth.service;

import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Service;

import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicLong;

/**
 * 用户凭据管理（内存实现，适用于本地开发与集成测试）。
 * <p>
 * 生产环境应接入 MySQL 等持久化存储。
 */
@Service
public class UserService {

    private static final String ADMIN_ID = "1";

    private final PasswordEncoder encoder = new BCryptPasswordEncoder();
    private final Map<String, User> byUsername = new ConcurrentHashMap<>();
    private final AtomicLong nextId = new AtomicLong(2);

    public UserService() {
        byUsername.put("admin", new User(ADMIN_ID, "admin", encoder.encode("admin123")));
    }

    public User authenticate(String username, String password) {
        User user = byUsername.get(username.trim().toLowerCase());
        if (user == null || !encoder.matches(password, user.passwordHash())) {
            throw new IllegalArgumentException("invalid credentials");
        }
        return user;
    }

    public User register(String username, String password) {
        String key = username.trim().toLowerCase();
        if (byUsername.containsKey(key)) {
            throw new IllegalStateException("username taken");
        }
        String id = String.valueOf(nextId.getAndIncrement());
        User user = new User(id, key, encoder.encode(password));
        byUsername.put(key, user);
        return user;
    }

    public record User(String id, String username, String passwordHash) {}
}
