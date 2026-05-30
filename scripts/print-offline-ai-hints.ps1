# UTF-8 hints after make offline-ai-up
$OutputEncoding = [Console]::OutputEncoding = [Text.UTF8Encoding]::new($false)
Write-Host ''
Write-Host 'Ollama:     http://localhost:11434/v1'
Write-Host 'OpenClaw:   http://localhost:18789  (token: .env.openclaw)'
Write-Host '设置 -> AI：选择 Ollama（本地离线），API 地址 http://127.0.0.1:11434/v1'
