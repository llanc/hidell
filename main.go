package main

import (
	_ "embed"
	"github.com/fsnotify/fsnotify"
	"github.com/getlantern/systray"
	"os"
)

//go:embed hidell.ico
var logo []byte

// 控制 watchDir goroutine 的通道
var watcherDone chan struct{}

// 全局 watcher 对象
var watcher *fsnotify.Watcher

func main() {
	LoadConfig()
	loadTranslations()
	watcherDone = make(chan struct{})

	// 初始化全局 watcher
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}

	go watchDir(getUserHome(), watcherDone)
	systray.Run(onReady, onExit)
}

func onExit() {
	// 清理资源
	if watcherDone != nil {
		close(watcherDone)
	}
	if watcher != nil {
		watcher.Close()
	}
}

func getUserHome() string {
	return os.Getenv("USERPROFILE")
}
