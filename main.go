package main

import (
	_ "embed"
	"github.com/getlantern/systray"
	"os"
)

//go:embed hidell.ico
var logo []byte

// 控制 watchDir goroutine 的通道
var watcherDone chan struct{}

func main() {
	LoadConfig()
	watcherDone = make(chan struct{})
	go watchDir(getUserHome(), watcherDone)
	systray.Run(onReady, onExit)
}

func onExit() {
	// 清理资源
	if watcherDone != nil {
		close(watcherDone)
	}
}

func getUserHome() string {
	return os.Getenv("USERPROFILE")
}
