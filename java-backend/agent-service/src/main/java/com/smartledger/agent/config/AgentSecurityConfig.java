package com.smartledger.agent.config;

import com.smartledger.agent.security.GatewayUserFilter;
import jakarta.servlet.DispatcherType;
import org.springframework.boot.context.properties.EnableConfigurationProperties;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.config.annotation.web.configuration.EnableWebSecurity;
import org.springframework.security.config.http.SessionCreationPolicy;
import org.springframework.security.web.SecurityFilterChain;
import org.springframework.security.web.authentication.UsernamePasswordAuthenticationFilter;

/**
 * Agent 服务安全：依赖网关 JWT 校验后注入的 {@code X-User-Id}，实现成员级访问隔离。
 */
@Configuration
@EnableWebSecurity
@EnableConfigurationProperties(AgentProperties.class)
public class AgentSecurityConfig {

    @Bean
    SecurityFilterChain agentSecurity(HttpSecurity http, GatewayUserFilter gatewayUserFilter) throws Exception {
        http.csrf(csrf -> csrf.disable())
                .sessionManagement(sm -> sm.sessionCreationPolicy(SessionCreationPolicy.STATELESS))
                .authorizeHttpRequests(auth -> auth
                        // SSE 异步派发/错误页不能再走鉴权，否则流式响应已提交后会 AccessDenied
                        .dispatcherTypeMatchers(DispatcherType.ASYNC, DispatcherType.ERROR).permitAll()
                        .requestMatchers("/api/v1/health").permitAll()
                        .requestMatchers("/api/v1/ai/**").authenticated()
                        .anyRequest().denyAll())
                .addFilterBefore(gatewayUserFilter, UsernamePasswordAuthenticationFilter.class);
        return http.build();
    }
}
