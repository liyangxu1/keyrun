<p align="right">
  <a href="./README.md">简体中文</a> | English
</p>

# KeyRun User Guide

This project is intended only to support students and researchers during the learning stage for study, research, and testing. Once you have the ability and proper conditions, please purchase and use an official JetBrains license to support high-quality tools and the broader software ecosystem.

JetBrains official website: [https://www.jetbrains.com/](https://www.jetbrains.com/)

KeyRun is a learning and research tool entry for JetBrains IDEs. It is used to quickly handle related JetBrains IDE configuration in a local test environment.

This tool is only for learning, research, and testing purposes. It must not be used for commercial scenarios, to bypass official software licensing, or for any purpose that violates software license agreements, platform rules, or local laws and regulations. For commercial use, please purchase an official JetBrains license.

## What It Does

- Provides operation entries for Windows, macOS, and Linux.
- Helps handle related startup configuration for locally installed JetBrains IDEs.
- Writes corresponding local configuration files for supported JetBrains products.
- Cleans up some historical configuration leftovers to reduce their impact on the current test environment.
- Provides a product list and quick search to confirm the currently covered product range.

Before running it, please understand that this tool modifies JetBrains IDE configuration under your local user directory. It is recommended to use it only in a virtual machine, test account, or rollback-ready environment.

## How To Use

1. Open the KeyRun page.
2. Close any running JetBrains IDE.
3. Choose the command for your operating system:
   - Windows: run it in an administrator PowerShell window.
   - macOS: run it in Terminal.
   - Linux: run it in Terminal.
4. Copy and run the command shown on the page.
5. Continue according to the script prompts.
6. Reopen your JetBrains IDE after the process is complete.

## Supported Products

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

## Before Use

- Windows requires running PowerShell with administrator privileges.
- macOS and Linux require an available terminal environment.
- macOS / Linux usually require `curl` and `jq`.
- Close all JetBrains IDEs before running it to avoid locked configuration files.
- If you have used other similar tools before, you may need to clean up old configuration and restart your computer first.
- This tool may affect existing IDE startup configuration. Back up important configuration before use.

## License

This project is licensed under the [GNU General Public License v3.0](./LICENSE).

## Anti-996 Statement

This project opposes the 996 work schedule. Users must keep the license notice, comply with applicable labor and employment laws, and must not force employees to waive their labor rights.

## Disclaimer

This tool is only for learning, research, and testing. It does not provide any commercial licensing capability and must not be used commercially or to bypass official licensing. Any licensing, compliance, data, configuration, or system risks caused by using this tool are the sole responsibility of the user.
