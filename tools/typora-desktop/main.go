package main

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

//go:embed runtime.js
var runtimeScript []byte

const marker = "KeyRun.Typora.Desktop.v1"

type installation struct {
	executable string
	archive    string
}

type backupState struct {
	User              string
	Original, Patched [sha256.Size]byte
	License           licenseSnapshot
}

func locate(dir string) (installation, error) {
	dirs := []string{dir}
	if dir == "" {
		dirs = candidates()
	}
	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		root, err := filepath.Abs(dir)
		if err != nil {
			continue
		}
		for _, name := range executableNames {
			exe := filepath.Join(root, name)
			archive := filepath.Join(root, "resources", "app.asar")
			info, exeErr := os.Stat(exe)
			pack, packErr := os.Lstat(archive)
			if exeErr == nil && info.Mode().IsRegular() && packErr == nil && pack.Mode().IsRegular() {
				return installation{exe, archive}, nil
			}
		}
	}
	return installation{}, fmt.Errorf("未找到 Typora，请用 --dir 指定包含 resources/app.asar 的安装目录")
}

func deviceFingerprint(id string) string {
	sum := sha256.Sum256([]byte(id + "typora"))
	return strings.NewReplacer("/", "a", "+", "a", "-", "a", "=", "a").Replace(base64.StdEncoding.EncodeToString(sum[:])[:10])
}

func inject(original []byte, version, id string) ([]byte, error) {
	if bytes.Contains(original, []byte(marker)) {
		return nil, fmt.Errorf("入口已被 KeyRun 处理")
	}
	if version == "" || id == "" {
		return nil, fmt.Errorf("缺少版本号或设备标识")
	}
	settings, err := json.Marshal(map[string]string{
		"marker": marker, "version": version,
		"original": base64.StdEncoding.EncodeToString(original), "fingerprint": deviceFingerprint(id),
	})
	if err != nil {
		return nil, err
	}
	out := append([]byte("\"use strict\";\n("), runtimeScript...)
	out = append(out, []byte(")(")...)
	out = append(out, settings...)
	out = append(out, []byte(");\n")...)
	return append(out, bytes.TrimPrefix(original, []byte{0xef, 0xbb, 0xbf})...), nil
}

func run() error {
	dir := flag.String("dir", "", "Typora 安装目录")
	status := flag.Bool("status", false, "只查看安装位置与备份状态")
	restore := flag.Bool("restore", false, "恢复原始应用包和许可证记录")
	flag.Parse()
	if flag.NArg() != 0 || (*status && *restore) {
		return fmt.Errorf("参数无效，请查看 --help")
	}
	app, err := locate(*dir)
	if err != nil {
		return err
	}
	backup := app.archive + ".keyrun-backup"
	state := backup + ".json"
	_, backupErr := os.Lstat(backup)
	_, stateErr := os.Lstat(state)
	if *status {
		fmt.Printf("应用：%s\n应用包：%s\n备份：%t\n记录备份：%t\n", app.executable, app.archive, backupErr == nil, stateErr == nil)
		return nil
	}
	running, err := typoraRunning()
	if err != nil {
		return err
	}
	if running {
		return fmt.Errorf("请保存文档并退出 Typora 后重试")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	info, err := os.Stat(app.archive)
	if err != nil {
		return err
	}
	if *restore {
		return restoreBackup(app, backup, state, home, info.Mode())
	}
	if !os.IsNotExist(backupErr) || !os.IsNotExist(stateErr) {
		return fmt.Errorf("备份已存在或不可访问：%s，请先检查备份", backup)
	}
	original, err := os.ReadFile(app.archive)
	if err != nil {
		return err
	}
	id, err := machineID()
	if err != nil {
		return err
	}
	patched, err := patchArchive(original, func(src []byte, _, version string) ([]byte, error) {
		return inject(src, version, id)
	})
	if err != nil {
		return err
	}
	record, err := readLicense()
	if err != nil {
		return err
	}
	stateData, err := json.Marshal(backupState{home, sha256.Sum256(original), sha256.Sum256(patched), record})
	if err != nil {
		return err
	}
	if err := writeInstallFile(backup, original, info.Mode(), true); err != nil {
		return err
	}
	if err := writeInstallFile(state, stateData, 0600, true); err != nil {
		return err
	}
	if err := writeInstallFile(app.archive, patched, info.Mode(), false); err != nil {
		return fmt.Errorf("写入应用包失败，备份已保留：%w", err)
	}
	date := time.Now().Format("1/2/2006")
	value := base64.StdEncoding.EncodeToString([]byte(marker)) + "#0#" + date
	if err := writeLicense(record, value, date); err != nil {
		rollback := restoreBackup(app, backup, state, home, info.Mode())
		if rollback != nil {
			return fmt.Errorf("写入记录失败：%v；恢复失败：%w", err, rollback)
		}
		return fmt.Errorf("写入记录失败，已恢复：%w", err)
	}
	fmt.Printf("已写入。请启动 Typora 检查状态。备份：%s\n", backup)
	return nil
}

func restoreBackup(app installation, backup, state, home string, mode os.FileMode) error {
	original, err := os.ReadFile(backup)
	if err != nil {
		return fmt.Errorf("无法读取应用包备份：%w", err)
	}
	data, err := readState(state)
	if err != nil {
		return fmt.Errorf("无法读取许可证记录备份：%w", err)
	}
	var snapshot backupState
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return err
	}
	if snapshot.User != home || snapshot.Original != sha256.Sum256(original) {
		return fmt.Errorf("备份与当前账户不匹配或备份内容已损坏")
	}
	current, err := os.ReadFile(app.archive)
	if err != nil {
		return err
	}
	switch sha256.Sum256(current) {
	case snapshot.Patched:
		if err := writeInstallFile(app.archive, original, mode, false); err != nil {
			return err
		}
	case snapshot.Original:
	default:
		return fmt.Errorf("当前应用包与备份不匹配，可能已经升级或重装")
	}
	if err := restoreLicense(snapshot.License); err != nil {
		return fmt.Errorf("应用包已恢复，许可证记录恢复失败：%w", err)
	}
	fmt.Printf("已恢复应用和许可证记录。备份保留在：%s\n", backup)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, "请保留完整错误信息，在项目 Issues 提交问题，或到 QQ 群 920344087 寻求帮助。")
		os.Exit(1)
	}
}
