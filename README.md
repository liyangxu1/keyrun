<p align="right">
  简体中文 | <a href="./README.en.md">English</a>
</p>

# KeyRun 使用说明

本项目仅用于支持学习阶段的学生和研究者进行学习、研究和测试。希望在具备条件和能力之后，优先购买并使用 JetBrains 正版授权，支持优秀工具和软件生态的持续发展。

JetBrains 官网：[https://www.jetbrains.com/](https://www.jetbrains.com/)

KeyRun 是一个面向 JetBrains 系列 IDE 的学习研究工具入口，用于在本机测试环境中快速处理 JetBrains IDE 的相关配置。

本工具仅用于学习、研究和测试用途。不得用于商业化场景，不得用于规避正版软件授权，不得用于任何违反软件许可协议、平台规则或当地法律法规的用途。商业使用请购买 JetBrains 官方许可证。

## 具体作用

- 为 Windows、macOS、Linux 提供对应的操作入口。
- 帮助处理本机已安装 JetBrains IDE 的相关启动配置。
- 为支持的 JetBrains 产品写入对应的本地配置文件。
- 清理部分历史配置残留，减少旧配置对当前测试环境的影响。
- 提供产品列表展示和快速搜索，方便确认当前覆盖的产品范围。

执行前请理解：本工具会修改本机用户目录下的 JetBrains IDE 配置。建议仅在虚拟机、测试账户或可回滚环境中使用。

## 怎么使用

1. 打开 KeyRun 页面。
2. 关闭正在运行的 JetBrains IDE。
3. 根据自己的系统选择对应命令：
   - Windows：使用 PowerShell 管理员窗口执行。
   - macOS：使用终端执行。
   - Linux：使用终端执行。
4. 复制页面中展示的对应命令并运行。
5. 按脚本提示继续操作。
6. 处理完成后重新打开 JetBrains IDE。

## 支持产品

- IntelliJ IDEA
- CLion
- PhpStorm
- GoLand
- PyCharm
- WebStorm
- Rider
- DataGrip
- RubyMine
- AppCode
- DataSpell
- RustRover

## 使用前注意

- Windows 需要使用管理员权限运行 PowerShell。
- macOS 和 Linux 需要可用的终端环境。
- macOS / Linux 环境通常需要 `curl` 和 `jq`。
- 执行前请关闭所有 JetBrains IDE，避免配置文件被占用。
- 如果之前使用过其他类似工具，可能需要先清理旧配置并重启电脑。
- 本工具可能影响已有 IDE 启动配置，使用前建议备份重要配置。

## 许可证

本项目采用 [GNU General Public License v3.0](./LICENSE) 开源协议。

## 免责声明

本工具仅供学习、研究和测试使用，不提供任何商业授权能力，不应被用于商业化使用或规避正版授权。由使用本工具产生的授权、合规、数据、配置或系统风险，均由使用者自行承担。
