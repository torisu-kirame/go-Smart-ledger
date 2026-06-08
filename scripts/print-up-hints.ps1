# UTF-8 hints after make up
$OutputEncoding = [Console]::OutputEncoding = [Text.UTF8Encoding]::new($false)
Write-Host ''
Write-Host 'Web UI:     http://localhost:25173'
Write-Host 'Gateway:    http://localhost:28080/api/v1/health'
Write-Host 'MiniLedger: http://localhost:24441/dashboard'
Write-Host 'Login:      admin / admin123'
Write-Host ''
Write-Host 'OpenClaw:   http://localhost:18789  (token: .env.openclaw)'
Write-Host ''
Write-Host 'AI：设置 -> AI -> API Key + Token -> 测试连接（Gateway 填 http://127.0.0.1:18789）'
Write-Host '离线 Ollama：make offline-ai-up'
