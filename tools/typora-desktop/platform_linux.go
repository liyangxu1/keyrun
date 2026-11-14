package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

var executableNames = []string{"Typora", "typora"}

func candidates() []string {
	dirs := []string{}
	if exe, err := exec.LookPath("typora"); err == nil {
		if resolved, err := filepath.EvalSymlinks(exe); err == nil {
			dirs = append(dirs, filepath.Dir(resolved))
		}
	}
	home, _ := os.UserHomeDir()
	return append(dirs, "/usr/share/typora", "/usr/lib/typora", "/usr/local/share/typora", "/opt/typora",
		filepath.Join(home, ".local", "share", "typora"))
}

func typoraRunning() (bool, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return false, err
	}
	for _, entry := range entries {
		if _, err := strconv.Atoi(entry.Name()); err != nil {
			continue
		}
		name, err := os.ReadFile(filepath.Join("/proc", entry.Name(), "comm"))
		if err == nil && strings.EqualFold(strings.TrimSpace(string(name)), "typora") {
			return true, nil
		}
	}
	return false, nil
}

func machineID() (string, error) {
	for _, path := range []string{"/etc/machine-id", "/var/lib/dbus/machine-id"} {
		if data, err := os.ReadFile(path); err == nil {
			if id := strings.TrimSpace(string(data)); id != "" {
				return id, nil
			}
		}
	}
	return "", fmt.Errorf("无法读取 Linux 设备标识")
}

type licenseSnapshot struct {
	Exists bool
	Data   []byte
}

func licensePath() (string, error) {
	if os.Geteuid() == 0 {
		return "", fmt.Errorf("请以桌面用户身份运行，由工具在写入安装目录时请求 sudo")
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	id, err := machineID()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "Typora", deviceFingerprint(id)), nil
}

func readState(name string) ([]byte, error) {
	data, err := os.ReadFile(name)
	if !os.IsPermission(err) {
		return data, err
	}
	cmd := exec.Command("sudo", "--", "cat", "--", name)
	cmd.Stdin, cmd.Stderr = os.Stdin, os.Stderr
	return cmd.Output()
}

func writeInstallFile(name string, data []byte, mode os.FileMode, exclusive bool) error {
	err := writeStaged(name, data, mode, exclusive)
	if !os.IsPermission(err) {
		return err
	}
	if _, err := exec.LookPath("sudo"); err != nil {
		return fmt.Errorf("安装目录不可写，且系统没有 sudo")
	}
	// 只有安装目录的文件操作需要提权，数据通过标准输入传入。
	cmd := exec.Command("sudo", "--", "/bin/sh", "-c", `
set -eu
temp=$(mktemp -- "${1%/*}/.keyrun-typora-XXXXXX")
trap 'rm -f -- "$temp"' EXIT
cat > "$temp"
chmod "$2" "$temp"
if [ "$3" = true ]; then
    ln -T -- "$temp" "$1"
else
    mv -fT -- "$temp" "$1"
fi
`, "keyrun", name, fmt.Sprintf("%04o", mode.Perm()), strconv.FormatBool(exclusive))
	cmd.Stdin = bytes.NewReader(data)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

func readLicense() (licenseSnapshot, error) {
	name, err := licensePath()
	if err != nil {
		return licenseSnapshot{}, err
	}
	info, err := os.Lstat(name)
	if os.IsNotExist(err) {
		return licenseSnapshot{}, nil
	}
	if err != nil {
		return licenseSnapshot{}, err
	}
	if !info.Mode().IsRegular() {
		return licenseSnapshot{}, fmt.Errorf("许可证记录不是普通文件")
	}
	data, err := os.ReadFile(name)
	return licenseSnapshot{Exists: true, Data: data}, err
}

func writeLicense(old licenseSnapshot, value, _ string) error {
	record := map[string]json.RawMessage{}
	if old.Exists {
		raw, err := hex.DecodeString(strings.TrimSpace(string(old.Data)))
		if err != nil {
			return err
		}
		if err := json.Unmarshal(raw, &record); err != nil || record == nil {
			return fmt.Errorf("无法读取现有许可证记录")
		}
	}
	token, _ := json.Marshal(value)
	record["SLicense"] = token
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}
	name, err := licensePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(name), 0700); err != nil {
		return err
	}
	return writeArchive(name, []byte(hex.EncodeToString(data)), 0600)
}

func restoreLicense(old licenseSnapshot) error {
	name, err := licensePath()
	if err != nil {
		return err
	}
	if old.Exists {
		return writeArchive(name, old.Data, 0600)
	}
	err = os.Remove(name)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
