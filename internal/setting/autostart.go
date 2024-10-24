package setting

import (
	"github.com/getlantern/systray"
	"golang.org/x/sys/windows/registry"
	"os"
	"os/exec"
)

func IsAutoStartEnabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()

	val, _, err := k.GetStringValue("HIDELL")
	if err != nil {
		return false
	}

	exePath, err := os.Executable()
	if err != nil {
		return false
	}

	return val == exePath
}

func EnableAutoStart() {
	exePath, err := os.Executable()
	if err != nil {
		return
	}

	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
	if err != nil {
		return
	}
	defer k.Close()

	_ = k.SetStringValue("HIDELL", exePath)
}

func DisableAutoStart() {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
	if err != nil {
		return
	}
	defer k.Close()

	_ = k.DeleteValue("HIDELL")
}

func RestartProgram() {
	executable, err := os.Executable()
	if err != nil {
		panic(err)
	}

	cmd := exec.Command(executable, os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err = cmd.Start()
	if err != nil {
		panic(err)
	}

	systray.Quit()
	os.Exit(0)
}
