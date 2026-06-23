param(
    [string] $Backend = "go"
)

$OutputEncoding = [Console]::OutputEncoding = [Text.UTF8Encoding]::new($false)
Write-Host ''
if ($Backend -eq 'java') {
    Write-Host 'Backend:    Java (java-backend/) + LangChain4j Agent'
    Write-Host 'Compose:    smart-ledger-java'
    Write-Host 'Containers: smart-ledger-java-*  (docker ps --filter name=smart-ledger-java)'
} else {
    Write-Host 'Backend:    Go (go-backend/) + py-backend AI'
    Write-Host 'Compose:    smart-ledger-go'
    Write-Host 'Containers: smart-ledger-go-*  (docker ps --filter name=smart-ledger-go)'
}
Write-Host 'Web UI:     http://localhost:25173'
Write-Host 'Gateway:    http://localhost:28080/api/v1/health'
Write-Host 'MiniLedger: http://localhost:24441/dashboard'
Write-Host 'Login:      admin / admin123'
Write-Host ''
if ($Backend -eq 'java') {
    Write-Host 'AI：设置 -> AI -> 配置 LLM（OpenAI 兼容 / Ollama）-> 测试连接'
    Write-Host '离线 Ollama：make offline-ai-up'
} else {
    Write-Host 'AI：设置 -> AI -> 配置 LLM -> 测试连接'
    Write-Host '离线 Ollama：make offline-ai-up'
}
Write-Host ''
