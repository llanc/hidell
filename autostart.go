package main

import (
	"golang.org/x/sys/windows/registry"
	"os"
)

func isAutoStartEnabled() bool {
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

func enableAutoStart() {
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

func disableAutoStart() {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
	if err != nil {
		return
	}
	defer k.Close()

	_ = k.DeleteValue("HIDELL")
}
