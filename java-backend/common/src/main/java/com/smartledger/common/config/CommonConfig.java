package com.smartledger.common.config;

import com.smartledger.common.jwt.JwtProperties;
import org.springframework.boot.context.properties.EnableConfigurationProperties;
import org.springframework.context.annotation.ComponentScan;
import org.springframework.context.annotation.Configuration;

/**
 * 各微服务通过 {@code @Import(CommonConfig.class)} 引入公共 Bean。
 * 避免在每个 Application 上重复写 ComponentScan。
 */
@Configuration
@ComponentScan("com.smartledger.common")
@EnableConfigurationProperties(JwtProperties.class)
public class CommonConfig {
}
