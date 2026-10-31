<p align="right">
  简体中文 | <a href="./README.en.md">English</a>
</p>

# KeyRun 使用说明

本项目仅用于支持学习阶段的学生和研究者进行学习、研究和测试。希望在具备条件和能力之后，优先购买并使用 JetBrains 正版授权，支持优秀工具和软件生态的持续发展。

JetBrains 官网：[https://www.jetbrains.com/](https://www.jetbrains.com/)

KeyRun 是一个面向 JetBrains 系列 IDE 的学习研究工具入口，用于在本机测试环境中快速处理 JetBrains IDE 的相关配置。

学习编程的小伙伴可以加入 QQ 群 `920344087`，有不懂的问题可以随时交流。

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

## Typora（macOS）

新增独立的 [Shell 工具](./scripts/typora.sh)，参考 [上游 mac.js](https://github.com/FlyAIBox/typora-activate-for-mac/blob/main/mac.js) 转换。使用 macOS 系统自带的 Bash、Perl、OpenSSL 等工具，无需 Node、npm、Python 或 Homebrew，也不会自动安装环境。与 JetBrains 脚本分别执行。

先保存文档并退出 Typora，在终端执行以下命令，下载成功后才运行；交互输入直接读取终端，支持机器码和邮箱输入：

```bash
(script=$(mktemp -t keyrun-typora) && trap 'rm -f "$script"' EXIT && curl -fL --retry 2 https://raw.githubusercontent.com/liyangxu1/keyrun/codex/typora-shell-tool/scripts/typora.sh -o "$script" && /bin/bash "$script")
```

也可下载后执行，自定义应用位置或恢复备份：

```bash
curl -fL --retry 2 https://raw.githubusercontent.com/liyangxu1/keyrun/codex/typora-shell-tool/scripts/typora.sh -o typora.sh
bash typora.sh --app "$HOME/Applications/Typora.app"
bash typora.sh --app "$HOME/Applications/Typora.app" --restore
```

- 默认路径为 `/Applications/Typora.app`；不要使用 `sudo`，应用和父目录需对当前用户可写。
- 脚本会替换二进制内的公钥及验证服务器地址，使用临时私钥生成离线 Token，并重新签名应用。
- 首次写入前备份完整应用至同级 `Typora.app.keyrun-backup`；写入或签名失败时尝试自动恢复，备份始终保留。已有备份时拒绝重复处理。
- Token 会显示在终端并尝试复制到剪贴板，需自行粘贴到 Typora 的离线激活页面。临时私钥、机器码及邮箱随脚本退出清理，不写入长期缓存。
- 仅支持二进制中包含上游原始公钥的版本，更新后不保证兼容；找不到公钥时终止且不修改应用。
- `--restore` 恢复备份的应用文件，不修改许可证偏好设置；不要在升级 Typora 后用旧版本备份覆盖新版本。
- 已通过临时文件测试验证字节长度、多个公钥替换、中文签名和异常输入，未对真实 Typora 执行激活验证。

## 许可证

本项目采用 [GNU General Public License v3.0](./LICENSE) 开源协议。

## 反 996 声明

本项目反对 996 工作制。使用者必须保留许可证声明，并且必须遵守劳动就业相关法律，不得强迫员工放弃劳动权益。

## 免责声明

本工具仅供学习、研究和测试使用，不提供任何商业授权能力，不应被用于商业化使用或规避正版授权。由使用本工具产生的授权、合规、数据、配置或系统风险，均由使用者自行承担。
