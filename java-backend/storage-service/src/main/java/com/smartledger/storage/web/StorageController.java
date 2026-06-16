package com.smartledger.storage.web;

import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.Map;

/** 备份与对象存储相关 API，当前仅提供健康检查 */
@RestController
@RequestMapping("/api/v1/storage")
public class StorageController {

    @GetMapping("/health")
    public Map<String, String> health() {
        return Map.of("status", "ok", "backend", "java");
    }
}
