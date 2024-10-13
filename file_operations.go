package main

import (
	"github.com/fsnotify/fsnotify"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

func hideDotFiles(dir string) {
	files, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if strings.HasPrefix(file.Name(), ".") {
			hideFile(filepath.Join(dir, file.Name()))
		}
	}
}

func unhideDotFiles(dir string) {
	files, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if strings.HasPrefix(file.Name(), ".") {
			unhideFile(filepath.Join(dir, file.Name()))
		}
	}
}

func hideFile(path string) {
	ptr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		panic(err)
	}
	attrs, err := syscall.GetFileAttributes(ptr)
	if err != nil {
		panic(err)
	}
	_ = syscall.SetFileAttributes(ptr, attrs|syscall.FILE_ATTRIBUTE_HIDDEN)
}

func unhideFile(path string) {
	ptr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		panic(err)
	}
	attrs, err := syscall.GetFileAttributes(ptr)
	if err != nil {
		panic(err)
	}
	_ = syscall.SetFileAttributes(ptr, attrs&^syscall.FILE_ATTRIBUTE_HIDDEN)
}

func watchDir(dir string, done chan struct{}) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	err = watcher.Add(dir)
	if err != nil {
		panic(err)
	}

	for {
		select {
		case <-done:
			return
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Create == fsnotify.Create {
				if strings.HasPrefix(filepath.Base(event.Name), ".") {
					hideFile(event.Name)
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			panic(err)
		}
	}
}
