# encoding: utf-8

param(
    [string]$PropertiesPath
)

# 此脚本只读取路径配置，不修改 IDE、系统环境变量或文件。
$OutputEncoding = [System.Text.Encoding]::UTF8
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$script:UserHome = [Environment]::GetFolderPath([Environment+SpecialFolder]::UserProfile)

function Resolve-IdeaPropertiesPath([string]$InputPath)
{
    if ([string]::IsNullOrWhiteSpace($InputPath))
    {
        return $null
    }

    $candidate = $InputPath.Trim().Trim('"')
    if (Test-Path -LiteralPath $candidate -PathType Leaf)
    {
        if ([System.IO.Path]::GetFileName($candidate) -ieq "idea.properties")
        {
            return (Resolve-Path -LiteralPath $candidate).Path
        }
        return $null
    }

    if (Test-Path -LiteralPath $candidate -PathType Container)
    {
        foreach ($relativePath in @("idea.properties", "bin\\idea.properties"))
        {
            $propertyFile = Join-Path -Path $candidate -ChildPath $relativePath
            if (Test-Path -LiteralPath $propertyFile -PathType Leaf)
            {
                return (Resolve-Path -LiteralPath $propertyFile).Path
            }
        }
    }

    return $null
}

function Add-Candidate([System.Collections.ArrayList]$Candidates, [string]$Path)
{
    $resolvedPath = Resolve-IdeaPropertiesPath $Path
    if ($null -ne $resolvedPath -and -not $Candidates.Contains($resolvedPath))
    {
        [void]$Candidates.Add($resolvedPath)
    }
}

function Get-IdeaPropertiesCandidates
{
    $candidates = New-Object System.Collections.ArrayList

    # IDEA_PROPERTIES 是 JetBrains 官方支持的自定义属性文件入口。
    foreach ($scope in @(
        [EnvironmentVariableTarget]::Process,
        [EnvironmentVariableTarget]::User,
        [EnvironmentVariableTarget]::Machine
    ))
    {
        Add-Candidate $candidates ([Environment]::GetEnvironmentVariable("IDEA_PROPERTIES", $scope))
    }

    $appData = [Environment]::GetFolderPath([Environment+SpecialFolder]::ApplicationData)
    $localAppData = [Environment]::GetFolderPath([Environment+SpecialFolder]::LocalApplicationData)
    $programFiles = @(
        [Environment]::GetFolderPath([Environment+SpecialFolder]::ProgramFiles),
        ${env:ProgramFiles(x86)}
    ) | Where-Object { -not [string]::IsNullOrWhiteSpace($_) }

    # 默认配置目录中可能存在通过“Edit Custom Properties”创建的属性文件。
    $defaultConfigRoot = Join-Path -Path $appData -ChildPath "JetBrains"
    if (Test-Path -LiteralPath $defaultConfigRoot -PathType Container)
    {
        Get-ChildItem -LiteralPath $defaultConfigRoot -Directory -ErrorAction SilentlyContinue | ForEach-Object {
            Add-Candidate $candidates (Join-Path -Path $_.FullName -ChildPath "idea.properties")
        }
    }

    # 只检查 JetBrains 的标准安装路径，不递归扫描用户磁盘。
    foreach ($programFilesRoot in $programFiles)
    {
        $jetBrainsRoot = Join-Path -Path $programFilesRoot -ChildPath "JetBrains"
        if (Test-Path -LiteralPath $jetBrainsRoot -PathType Container)
        {
            Get-ChildItem -LiteralPath $jetBrainsRoot -Directory -ErrorAction SilentlyContinue | ForEach-Object {
                Add-Candidate $candidates $_.FullName
            }
        }
    }

    $programsRoot = Join-Path -Path $localAppData -ChildPath "Programs\\JetBrains"
    if (Test-Path -LiteralPath $programsRoot -PathType Container)
    {
        Get-ChildItem -LiteralPath $programsRoot -Directory -ErrorAction SilentlyContinue | ForEach-Object {
            Add-Candidate $candidates $_.FullName
        }
    }

    $toolboxRoot = Join-Path -Path $localAppData -ChildPath "JetBrains\\Toolbox\\apps"
    if (Test-Path -LiteralPath $toolboxRoot -PathType Container)
    {
        Get-ChildItem -LiteralPath $toolboxRoot -Filter "idea.properties" -File -Recurse -ErrorAction SilentlyContinue | ForEach-Object {
            Add-Candidate $candidates $_.FullName
        }
    }

    # 保持 ArrayList 类型，避免 PowerShell 将单个候选项自动展开为字符串。
    return ,$candidates
}

function Resolve-ConfiguredPath([string]$Value, [string]$PropertyFilePath)
{
    if ([string]::IsNullOrWhiteSpace($Value))
    {
        return $null
    }

    $expandedValue = [Environment]::ExpandEnvironmentVariables($Value.Trim())
    $expandedValue = $expandedValue.Replace('${user.home}', $script:UserHome).Replace('/', '\')

    if (-not [System.IO.Path]::IsPathRooted($expandedValue))
    {
        $propertiesDirectory = Split-Path -Path $PropertyFilePath -Parent
        $expandedValue = Join-Path -Path $propertiesDirectory -ChildPath $expandedValue
    }

    try
    {
        return [System.IO.Path]::GetFullPath($expandedValue)
    }
    catch
    {
        return $expandedValue
    }
}

function Get-IdeaPropertyValues([string]$PropertyFilePath)
{
    $properties = @{}
    Get-Content -LiteralPath $PropertyFilePath -Encoding UTF8 -ErrorAction Stop | ForEach-Object {
        $line = $_.Trim()
        if ($line -and -not $line.StartsWith("#") -and $line -match "^\\s*([^#=]+?)\\s*=\\s*(.*)$")
        {
            $properties[$matches[1].Trim()] = $matches[2].Trim()
        }
    }
    return $properties
}

function Show-PathStatus([string]$Name, [string]$Path)
{
    if ([string]::IsNullOrWhiteSpace($Path))
    {
        Write-Host ("{0}: 未设置，IDE 将使用默认目录。" -f $Name) -ForegroundColor DarkYellow
        return
    }

    $status = if (Test-Path -LiteralPath $Path) { "存在" } else { "不存在" }
    $color = if ($status -eq "存在") { "Green" } else { "Yellow" }
    Write-Host ("{0}: {1} ({2})" -f $Name, $Path, $status) -ForegroundColor $color
}

function Invoke-IdeaPathDiagnosis
{
    $propertyFile = Resolve-IdeaPropertiesPath $PropertiesPath
    if ($null -eq $propertyFile)
    {
        $candidates = Get-IdeaPropertiesCandidates
        if ($candidates.Count -eq 1)
        {
            $propertyFile = $candidates[0]
        }
        elseif ($candidates.Count -gt 1)
        {
            Write-Host "找到多个 idea.properties 文件，请使用 -PropertiesPath 指定其中一个：" -ForegroundColor Yellow
            $candidates | ForEach-Object { Write-Host "  $_" }
            return
        }
    }

    if ($null -eq $propertyFile)
    {
        Write-Host "未在官方默认目录或标准安装目录中找到 idea.properties。" -ForegroundColor Yellow
        $manualPath = Read-Host "请输入 idea.properties 完整路径或 IDE 安装目录"
        $propertyFile = Resolve-IdeaPropertiesPath $manualPath
    }

    if ($null -eq $propertyFile)
    {
        Write-Error "未找到有效的 idea.properties。请确认 IDE 已按标准方式安装，或提供正确路径。"
        return
    }

    try
    {
        $properties = Get-IdeaPropertyValues $propertyFile
    }
    catch
    {
        Write-Error ("读取 idea.properties 失败：{0}" -f $_.Exception.Message)
        return
    }

    $configPath = Resolve-ConfiguredPath $properties["idea.config.path"] $propertyFile
    $systemPath = Resolve-ConfiguredPath $properties["idea.system.path"] $propertyFile

    Write-Host "`nJetBrains 配置路径诊断结果" -ForegroundColor Cyan
    Write-Host ("idea.properties: {0}" -f $propertyFile)
    Show-PathStatus "idea.config.path" $configPath
    Show-PathStatus "idea.system.path" $systemPath

    if ($configPath -and $systemPath -and $configPath -ieq $systemPath)
    {
        Write-Warning "idea.config.path 与 idea.system.path 不能使用同一目录。"
    }
}

Invoke-IdeaPathDiagnosis
