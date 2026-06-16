package com.smartledger.ledger.web;

import com.smartledger.common.web.ApiErrorAdvice.ApiException;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

/** AI 助手接口占位，功能开发中 */
@RestController
@RequestMapping("/api/v1/ai")
public class AiStubController {

    @RequestMapping("/**")
    public void notImplemented() {
        throw ApiException.notImplemented("AI assistant is not available in this build yet");
    }
}
