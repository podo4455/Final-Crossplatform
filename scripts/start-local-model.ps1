$ErrorActionPreference = 'Stop'
$projectRoot = Split-Path -Parent $PSScriptRoot
$portable = Join-Path $projectRoot '.tools/ollama/ollama.exe'
$env:OLLAMA_HOST = '127.0.0.1:11434'
$env:OLLAMA_NO_CLOUD = '1'
if (Test-Path -LiteralPath $portable) {
    $ollamaCommand = $portable
    $env:OLLAMA_MODELS = Join-Path $projectRoot '.tools/ollama/models'
} else {
    $ollamaCommand = (Get-Command ollama -ErrorAction Stop).Source
}
$ready = $false
try {
    $null = Invoke-RestMethod 'http://127.0.0.1:11434/api/tags' -TimeoutSec 2
    $ready = $true
} catch {}
if (-not $ready) {
    Start-Process -FilePath $ollamaCommand -ArgumentList 'serve' -WindowStyle Hidden
    for ($attempt = 0; $attempt -lt 30; $attempt++) {
        Start-Sleep -Seconds 1
        try {
            $null = Invoke-RestMethod 'http://127.0.0.1:11434/api/tags' -TimeoutSec 2
            $ready = $true
            break
        } catch {}
    }
}
if (-not $ready) { throw 'Ollama did not start on 127.0.0.1:11434' }
& $ollamaCommand pull gemma3:4b
if ($LASTEXITCODE -ne 0) { throw 'Could not prepare gemma3:4b' }
Write-Output 'Local image model is ready. Open Photo Backup and choose Gallery.'
