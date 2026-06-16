package com.smartledger.common.web;

import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.RestControllerAdvice;

import java.util.Map;

/**
 * 统一错误响应：{@code {"code":4xx,"msg":"..."}}。
 * <p>
 * 与前端 {@code ApiError} 解析约定一致。
 */
@RestControllerAdvice
public class ApiErrorAdvice {

    @ExceptionHandler(ApiException.class)
    public ResponseEntity<Map<String, Object>> handleApi(ApiException ex) {
        return ResponseEntity.status(ex.status())
                .body(Map.of("code", ex.status(), "msg", ex.getMessage()));
    }

    /** 业务层可预期 HTTP 错误，由 Advice 序列化为 JSON */
    public static class ApiException extends RuntimeException {
        private final int status;

        public ApiException(int status, String message) {
            super(message);
            this.status = status;
        }

        public int status() {
            return status;
        }

        public static ApiException badRequest(String msg) {
            return new ApiException(HttpStatus.BAD_REQUEST.value(), msg);
        }

        public static ApiException unauthorized(String msg) {
            return new ApiException(HttpStatus.UNAUTHORIZED.value(), msg);
        }

        public static ApiException notImplemented(String msg) {
            return new ApiException(HttpStatus.NOT_IMPLEMENTED.value(), msg);
        }
    }
}
