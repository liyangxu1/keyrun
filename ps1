# Bootstrap script - downloads and executes main script with correct UTF-8 encoding
chcp 65001 | Out-Null
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$bytes = (Invoke-WebRequest -Uri "https://liyangxu1.github.io/keyrun/scripts/install.ps1" -UseBasicParsing).Content
Invoke-Expression ([System.Text.Encoding]::UTF8.GetString($bytes))
