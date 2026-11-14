package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var executableNames = []string{"Typora.exe"}

const licenseKey = `HKCU\SOFTWARE\Typora`

func readState(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func writeInstallFile(name string, data []byte, mode os.FileMode, exclusive bool) error {
	err := writeStaged(name, data, mode, exclusive)
	if os.IsPermission(err) {
		return fmt.Errorf("安装目录不可写，请使用当前账户的管理员 PowerShell：%w", err)
	}
	return err
}

func queryValue(key, name string) (string, bool, error) {
	args := []string{"query", key}
	if name == "" {
		args = append(args, "/ve")
	} else {
		args = append(args, "/v", name)
	}
	out, err := exec.Command("reg.exe", args...).Output()
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok && e.ExitCode() == 1 {
			return "", false, nil
		}
		return "", false, err
	}
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.SplitN(strings.TrimSuffix(line, "\r"), "    REG_SZ", 2)
		if len(fields) == 2 {
			return strings.TrimPrefix(fields[1], "    "), true, nil
		}
	}
	return "", false, fmt.Errorf("无法解析注册表字符串：%s", key)
}

func candidates() []string {
	var dirs []string
	for _, hive := range []string{`HKCU\SOFTWARE`, `HKLM\SOFTWARE`, `HKLM\SOFTWARE\WOW6432Node`} {
		exe, ok, _ := queryValue(hive+`\Microsoft\Windows\CurrentVersion\App Paths\Typora.exe`, "")
		if ok {
			dirs = append(dirs, filepath.Dir(strings.Trim(exe, "\"")))
		}
	}
	if exe, err := exec.LookPath("Typora.exe"); err == nil {
		dirs = append(dirs, filepath.Dir(exe))
	}
	for _, base := range []string{os.Getenv("ProgramFiles"), os.Getenv("ProgramFiles(x86)")} {
		if base != "" {
			dirs = append(dirs, filepath.Join(base, "Typora"))
		}
	}
	if base := os.Getenv("LOCALAPPDATA"); base != "" {
		dirs = append(dirs, filepath.Join(base, "Programs", "Typora"))
	}
	return dirs
}

func typoraRunning() (bool, error) {
	out, err := exec.Command("tasklist.exe", "/FI", "IMAGENAME eq Typora.exe", "/FO", "CSV", "/NH").Output()
	return strings.Contains(strings.ToLower(string(out)), "\"typora.exe\""), err
}

func machineID() (string, error) {
	id, ok, err := queryValue(`HKLM\SOFTWARE\Microsoft\Cryptography`, "MachineGuid")
	if err != nil {
		return "", err
	}
	if !ok || id == "" {
		return "", fmt.Errorf("无法读取 Windows 设备标识")
	}
	return id, nil
}

type licenseSnapshot struct {
	License *string
	Date    *string
}

func readLicense() (licenseSnapshot, error) {
	var record licenseSnapshot
	for name, target := range map[string]**string{"SLicense": &record.License, "IDate": &record.Date} {
		value, exists, err := queryValue(licenseKey, name)
		if err != nil {
			return licenseSnapshot{}, err
		}
		if exists {
			*target = &value
		}
	}
	return record, nil
}

func setValue(name string, value *string) error {
	if value == nil {
		_, exists, err := queryValue(licenseKey, name)
		if err != nil || !exists {
			return err
		}
		return exec.Command("reg.exe", "delete", licenseKey, "/v", name, "/f").Run()
	}
	return exec.Command("reg.exe", "add", licenseKey, "/v", name, "/t", "REG_SZ", "/d", *value, "/f").Run()
}

func writeLicense(_ licenseSnapshot, value, date string) error {
	if err := setValue("SLicense", &value); err != nil {
		return err
	}
	return setValue("IDate", &date)
}

func restoreLicense(record licenseSnapshot) error {
	if err := setValue("SLicense", record.License); err != nil {
		return err
	}
	return setValue("IDate", record.Date)
}
