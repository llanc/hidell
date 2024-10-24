package main

import (
	_ "embed"
	"github.com/getlantern/systray"
	"hidell/internal/bootstrap"
	"hidell/internal/global"
	"hidell/internal/gui"
)

func main() {
	bootstrap.Init()
	systray.Run(gui.SystrayInit, onExit)

}

func onExit() {
	if global.Watcher != nil {
		_ = global.Watcher.Close()
	}
}
