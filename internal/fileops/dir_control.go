package fileops

import (
	"github.com/fsnotify/fsnotify"
	"github.com/getlantern/systray"
	"github.com/ncruces/zenity"
	"hidell/configs"
	"hidell/internal/base"
	"hidell/internal/global"
	"hidell/internal/local"
	"path/filepath"
	"strings"
)

type DirCtrl struct {
	Dir  configs.CustomDirs
	Done chan struct{}
}

func MakeDir(dir *configs.CustomDirs) DirCtrl {
	return DirCtrl{
		Dir: configs.CustomDirs{
			Path:   dir.Path,
			Alias:  dir.Alias,
			Active: dir.Active,
			Hidden: dir.Hidden,
		},
		Done: make(chan struct{}),
	}
}

func MakeDirDefault(path, alias string) DirCtrl {
	return DirCtrl{
		Dir: configs.CustomDirs{
			Path:   path,
			Alias:  alias,
			Active: true,
			Hidden: true,
		},
		Done: make(chan struct{}),
	}
}

func AddDir(parentMenu *systray.MenuItem) {
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
		local.T("optional"),
		zenity.Title(local.T("setAlias")),
		zenity.CancelLabel(local.T("cancel")),
		zenity.OKLabel(local.T("ok")),
		zenity.WindowIcon("assets/hidell.ico"),
		zenity.Attach(base.GetForegroundWindow()),
	)

	if err != nil {
		if err == zenity.ErrCanceled {
			alias = ""
		} else {
			zenity.Error("输入别名时出错: " + err.Error())
			return
		}
	}

	newDir := MakeDirDefault(path, alias)
	//dirs = append(dirs, newDir)
	newDir.AddToMenu(parentMenu)
	HideDotDir(path)
	//dirs[len(dirs)-1].StartWatching()

	global.Conf.CustomDirs = append(global.Conf.CustomDirs, configs.CustomDirs{
		Path:   newDir.Dir.Path,
		Alias:  newDir.Dir.Alias,
		Active: newDir.Dir.Active,
		Hidden: newDir.Dir.Hidden,
	})

	base.SaveConfig()
}

func (dirCtrl *DirCtrl) AddToMenu(parentMenu *systray.MenuItem) {
	menuName := dirCtrl.Dir.Alias
	if menuName == "" {
		menuName = dirCtrl.Dir.Path
	}
	dirMenu := parentMenu.AddSubMenuItem(menuName, "")
	mActivate := dirMenu.AddSubMenuItem(local.T("activate"), "")
	mShow := dirMenu.AddSubMenuItem(local.T("show"), "")
	mHide := dirMenu.AddSubMenuItem(local.T("hide"), "")
	mRemove := dirMenu.AddSubMenuItem(local.T("remove"), "")

	if dirCtrl.Dir.Active {
		mActivate.Check()
		dirCtrl.StartWatching()
	}
	if dirCtrl.Dir.Hidden {
		mHide.Check()
	} else {
		mShow.Check()
	}

	go func() {
		for {
			select {
			case <-mActivate.ClickedCh:
				dirCtrl.ToggleActivate(mActivate)
			case <-mShow.ClickedCh:
				dirCtrl.Show(mShow, mHide)
			case <-mHide.ClickedCh:
				dirCtrl.Hide(mShow, mHide)
			case <-mRemove.ClickedCh:
				dirCtrl.remove(dirMenu)
				return
			}
		}
	}()
}

func (dirCtrl *DirCtrl) ToggleActivate(menuItem *systray.MenuItem) {
	dirCtrl.Dir.Active = !dirCtrl.Dir.Active
	if dirCtrl.Dir.Active {
		menuItem.Check()
		dirCtrl.StartWatching()
	} else {
		menuItem.Uncheck()
		dirCtrl.StopWatching()
	}
}

func (dirCtrl *DirCtrl) Show(showItem, hideItem *systray.MenuItem) {
	dirCtrl.Dir.Hidden = false
	showItem.Check()
	hideItem.Uncheck()
	ShowDotDir(dirCtrl.Dir.Path)
	base.SaveConfig()
}

func (dirCtrl *DirCtrl) Hide(showItem, hideItem *systray.MenuItem) {
	dirCtrl.Dir.Hidden = true
	hideItem.Check()
	showItem.Uncheck()
	HideDotDir(dirCtrl.Dir.Path)
	base.SaveConfig()
}

func (dirCtrl *DirCtrl) remove(dirMenu *systray.MenuItem) {
	dirCtrl.StopWatching()
	for i, d := range global.Conf.CustomDirs {
		if d.Path == dirCtrl.Dir.Path {
			global.Conf.CustomDirs = append(global.Conf.CustomDirs[:i], global.Conf.CustomDirs[i+1:]...)
			break
		}
	}
	dirMenu.Hide()
	base.SaveConfig()
}

func (dirCtrl *DirCtrl) StartWatching() {
	if global.Watcher == nil {
		return
	}
	dirCtrl.Done = make(chan struct{})

	go func() {
		dirCtrl.watchDir()
	}()

	err := global.Watcher.Add(dirCtrl.Dir.Path)
	if err != nil {
		_ = zenity.Error(err.Error())
		dirCtrl.StopWatching()
	}
}

func (dirCtrl *DirCtrl) StopWatching() {
	if global.Watcher == nil {
		return
	}
	select {
	case <-dirCtrl.Done:
		// 通道已关闭，无需再次关闭
	default:
		close(dirCtrl.Done)
	}
	_ = global.Watcher.Remove(dirCtrl.Dir.Path)
}

func (dirCtrl *DirCtrl) watchDir() {
	for {
		select {
		case <-dirCtrl.Done:
			return
		case event, ok := <-global.Watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Create == fsnotify.Create {
				if strings.HasPrefix(filepath.Base(event.Name), ".") {
					HideFile(event.Name)
				}
			}
		case err, ok := <-global.Watcher.Errors:
			if !ok {
				return
			}
			zenity.Error("监视目录时出错: " + err.Error())
		}
	}
}
