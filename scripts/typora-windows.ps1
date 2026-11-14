param(
    [string]$Dir,
    [switch]$Status,
    [switch]$Restore
)

$ErrorActionPreference = 'Stop'
Write-Host @'

 _  __          ____
| |/ /___ _   _|  _ \ _   _ _ __
| ' // _ \ | | | |_) | | | | '_ \
| . \  __/ |_| |  _ <| |_| | | | |
|_|\_\___|\__, |_| \_\\__,_|_| |_|
          |___/
'@ -ForegroundColor Cyan
$author = [string][char]0x9152 + [char]0x6C8F + [char]0x8336
$tool = -join ([char]0x5DE5, [char]0x5177)
Write-Host "      Typora $tool | KeyRun" -ForegroundColor White
Write-Host "                        By $author`n" -ForegroundColor DarkGray

$release = 'https://github.com/liyangxu1/keyrun/releases/download/typora-desktop'
$work = Join-Path ([IO.Path]::GetTempPath()) ("keyrun-typora-" + [guid]::NewGuid())
try {
    $arch = $env:PROCESSOR_ARCHITECTURE
    if ($env:PROCESSOR_ARCHITEW6432) { $arch = $env:PROCESSOR_ARCHITEW6432 }
    $asset = switch ($arch) {
        'AMD64' { 'typora-windows-amd64.exe' }
        'ARM64' { 'typora-windows-arm64.exe' }
        default { throw "不支持的 Windows 架构：$arch" }
    }
    New-Item -ItemType Directory -Path $work | Out-Null
    $binary = Join-Path $work $asset
    $manifest = Join-Path $work 'SHA256SUMS'
    Invoke-WebRequest -Uri "$release/$asset" -UseBasicParsing -OutFile $binary
    Invoke-WebRequest -Uri "$release/SHA256SUMS" -UseBasicParsing -OutFile $manifest
    $pattern = '^([a-fA-F0-9]{64})\s+\*?' + [regex]::Escape($asset) + '$'
    $entry = @(Select-String -LiteralPath $manifest -Pattern $pattern)
    if ($entry.Count -ne 1 -or $entry[0].Matches[0].Groups[1].Value -ne (Get-FileHash -LiteralPath $binary -Algorithm SHA256).Hash) {
        throw '下载文件的 SHA-256 校验失败。'
    }
    $cliArgs = @()
    if ($Dir) { $cliArgs += @('--dir', $Dir) }
    if ($Status) { $cliArgs += '--status' }
    if ($Restore) { $cliArgs += '--restore' }
    & $binary @cliArgs
    if ($LASTEXITCODE -ne 0) { throw "Typora 工具退出码：$LASTEXITCODE" }
}
finally {
    if (Test-Path -LiteralPath $work) { Remove-Item -LiteralPath $work -Recurse -Force }
}
