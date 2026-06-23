package com.smartledger.common.config;

import com.smartledger.common.jwt.JwtProperties;
import com.smartledger.common.jwt.JwtService;
import org.springframework.boot.context.properties.EnableConfigurationProperties;
import org.springframework.context.annotation.ComponentScan;
import org.springframework.context.annotation.Configuration;

/** 网关等 Reactive 服务仅需 JWT，勿引入 Servlet MVC 组件 */
@Configuration
@ComponentScan(basePackageClasses = JwtService.class)
@EnableConfigurationProperties(JwtProperties.class)
public class CommonJwtConfig {
}
