package main

import (
	"github.com/fsnotify/fsnotify"
	"github.com/getlantern/systray"
	"github.com/ncruces/zenity"
	"path/filepath"
	"strings"
)

type CustomDir struct {
	Path    string `json:"path"`
	Alias   string `json:"alias"`
	Active  bool   `json:"active"`
	Hidden  bool   `json:"hidden"`
	watcher *fsnotify.Watcher
	done    chan struct{}
}

var customDirs []CustomDir

func addCustomDirMenu(dir *CustomDir, parentMenu *systray.MenuItem) {
	menuName := dir.Alias
	if menuName == "" {
		menuName = dir.Path
	}
	dirMenu := parentMenu.AddSubMenuItem(menuName, "")
	mActivate := dirMenu.AddSubMenuItem(t("activate"), "")
	mShow := dirMenu.AddSubMenuItem(t("show"), "")
	mHide := dirMenu.AddSubMenuItem(t("hide"), "")
	mRemove := dirMenu.AddSubMenuItem(t("remove"), "")

	if dir.Active {
		mActivate.Check()
		dir.startWatching()
	}
	if dir.Hidden {
		mHide.Check()
	} else {
		mShow.Check()
	}

	go func() {
		for {
			select {
			case <-mActivate.ClickedCh:
				toggleActivate(dir, mActivate)
			case <-mShow.ClickedCh:
				showCustomDir(dir, mShow, mHide)
			case <-mHide.ClickedCh:
				hideCustomDir(dir, mShow, mHide)
			case <-mRemove.ClickedCh:
				removeCustomDir(dir, dirMenu, parentMenu)
				return
			}
		}
	}()
}

func addNewCustomDir(parentMenu *systray.MenuItem) {
	path, err := zenity.SelectFile(
		zenity.Title("选择目录"),
		zenity.Directory(),
	)
	if err != nil {
		if err == zenity.ErrCanceled {
			return
		}
		zenity.Error("选择目录时出错: " + err.Error())
		return
	}

	alias, err := zenity.Entry(
		"为这个目录输入一个别名（可选）",
		zenity.Title("设置别名"),
	)
	if err != nil {
		if err == zenity.ErrCanceled {
			alias = ""
		} else {
			zenity.Error("输入别名时出错: " + err.Error())
			return
		}
	}

	newDir := CustomDir{
		Path:    path,
		Alias:   alias,
		Active:  true,
		Hidden:  true,
		watcher: nil,
		done:    make(chan struct{}),
	}
	customDirs = append(customDirs, newDir)
	addCustomDirMenu(&customDirs[len(customDirs)-1], parentMenu)
	hideDotFiles(path)
	customDirs[len(customDirs)-1].startWatching()
	saveConfig()
}

func toggleActivate(dir *CustomDir, menuItem *systray.MenuItem) {
	dir.Active = !dir.Active
	if dir.Active {
		menuItem.Check()
		dir.startWatching()
	} else {
		menuItem.Uncheck()
		dir.stopWatching()
	}
	saveConfig()
}

func showCustomDir(dir *CustomDir, showItem, hideItem *systray.MenuItem) {
	dir.Hidden = false
	showItem.Check()
	hideItem.Uncheck()
	unhideDotFiles(dir.Path)
	saveConfig()
}

func hideCustomDir(dir *CustomDir, showItem, hideItem *systray.MenuItem) {
	dir.Hidden = true
	hideItem.Check()
	showItem.Uncheck()
	hideDotFiles(dir.Path)
	saveConfig()
}

func removeCustomDir(dir *CustomDir, dirMenu, parentMenu *systray.MenuItem) {
	dir.stopWatching()
	for i, d := range customDirs {
		if d.Path == dir.Path {
			customDirs = append(customDirs[:i], customDirs[i+1:]...)
			break
		}
	}
	dirMenu.Hide()
	saveConfig()
}

func (dir *CustomDir) startWatching() {
	if dir.watcher != nil {
		return
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		zenity.Error("创建监视器时出错: " + err.Error())
		return
	}

	dir.watcher = watcher
	dir.done = make(chan struct{})

	go func() {
		defer watcher.Close()
		dir.watchDir()
	}()

	err = watcher.Add(dir.Path)
	if err != nil {
		zenity.Error("添加监视目录时出错: " + err.Error())
		dir.stopWatching()
	}
}

func (dir *CustomDir) stopWatching() {
	if dir.watcher == nil {
		return
	}

	close(dir.done)
	dir.watcher.Close()
	dir.watcher = nil
}

func (dir *CustomDir) watchDir() {
	for {
		select {
		case <-dir.done:
			return
		case event, ok := <-dir.watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Create == fsnotify.Create {
				if strings.HasPrefix(filepath.Base(event.Name), ".") {
					hideFile(event.Name)
				}
			}
		case err, ok := <-dir.watcher.Errors:
			if !ok {
				return
			}
			zenity.Error("监视目录时出错: " + err.Error())
		}
	}
}
