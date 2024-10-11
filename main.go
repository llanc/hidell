package main

import (
	_ "embed"
	"encoding/json"
	"github.com/fsnotify/fsnotify"
	"github.com/getlantern/systray"
	"github.com/ncruces/zenity"
	"github.com/skratchdot/open-golang/open"
	"golang.org/x/sys/windows/registry"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

//go:embed hidell.ico
var logo []byte

var systrayToolTip = "HIDE Like Linux"

// 控制 watchDir goroutine 的通道
var watcherDone chan struct{}

// 添加新的结构体来表示自定义目录
type CustomDir struct {
	Path    string `json:"path"`
	Alias   string `json:"alias"`
	Active  bool   `json:"active"`
	Hidden  bool   `json:"hidden"`
	watcher *fsnotify.Watcher
	done    chan struct{}
}

// 添加全局变量来存储自定义目录
var customDirs []CustomDir

func main() {
	loadConfig()
	//启动快捷键监听
	//go func() {
	//	hkey := hotkey.New()
	//	hkey.Register(hotkey.Ctrl, 'H', func() {
	//		fmt.Println("Push Ctrl+H")
	//	})
	//}()

	watcherDone = make(chan struct{})
	go watchDir(os.Getenv("USERPROFILE"), watcherDone)
	//运行托盘菜单
	systray.Run(onReady, onExit)

}

func onReady() {
	systray.SetIcon(logo)
	systray.SetTitle("HIDELL")
	systray.SetTooltip(systrayToolTip)

	// 定义托盘图标的右键菜单
	//runWindow := systray.AddMenuItem("主界面", "")
	//systray.AddSeparator()
	//mAdd := systray.AddMenuItem("添加目录", "")
	//systray.AddSeparator()
	mMonitor := systray.AddMenuItem("激活", "自动隐藏新生成的点文件/点文件夹")
	mMonitor.Check()
	systray.AddSeparator()
	mHide := systray.AddMenuItem("隐藏", "隐藏已存在的点文件/点文件夹")
	mShow := systray.AddMenuItem("显示", "显示已存在的点文件/点文件夹")
	systray.AddSeparator()

	mOtherDirs := systray.AddMenuItem("其他目录", "管理其他目录")

	systray.AddSeparator()
	mAbout := systray.AddMenuItem("关于HIDELL V1.2", "")
	systray.AddSeparator()
	mAutoStart := systray.AddMenuItem("开机自启", "设置程序开机自启")
	if isAutoStartEnabled() {
		mAutoStart.Check()
	}
	mQuit := systray.AddMenuItem("退出", "")

	mAddDir := mOtherDirs.AddSubMenuItem("添加目录", "添加新的目录")
	// 添加自定义目录菜单
	for i := range customDirs {
		addCustomDirMenu(&customDirs[i], mOtherDirs)
	}

	userHome := os.Getenv("USERPROFILE")
	go func() {
		for {
			select {
			//case <-runWindow.ClickedCh:
			//	go func() {
			//		runHomeWindow()
			//	}()
			//case <-mAdd.ClickedCh:
			//	// TODO: 打开文件夹选择器
			case <-mMonitor.ClickedCh:
				if mMonitor.Checked() {
					mMonitor.Uncheck() // 取消选中
					if watcherDone != nil {
						close(watcherDone) // 发送停止信号
						watcherDone = nil
					}
				} else {
					mMonitor.Check() // 选中
					watcherDone = make(chan struct{})
					go watchDir(userHome, watcherDone)
				}
			case <-mHide.ClickedCh:
				mHide.Check()
				mShow.Uncheck()
				hideDotFiles(userHome)
			case <-mShow.ClickedCh:
				mShow.Check()
				mHide.Uncheck()
				unhideDotFiles(userHome)
			case <-mAutoStart.ClickedCh:
				if mAutoStart.Checked() {
					disableAutoStart()
					mAutoStart.Uncheck()
				} else {
					enableAutoStart()
					mAutoStart.Check()
				}
			case <-mAbout.ClickedCh:
				_ = open.Run("https://www.github.com/llanc/hidell")
			case <-mQuit.ClickedCh:
				systray.Quit() //退出托盘
				return
			case <-mAddDir.ClickedCh:
				addNewCustomDir(mOtherDirs)
			}
		}
	}()

}

func onExit() {
	// 清理资源
	if watcherDone != nil {
		close(watcherDone)
	}
}

// 隐藏目录下所有以点开头的文件或文件夹
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

// 取消隐藏目录下所有以点开头的文件或文件夹
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

// 隐藏文件
func hideFile(path string) {
	ptr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		panic(err)
		return
	}
	attrs, err := syscall.GetFileAttributes(ptr)
	if err != nil {
		panic(err)
		return
	}
	_ = syscall.SetFileAttributes(ptr, attrs|syscall.FILE_ATTRIBUTE_HIDDEN)
}

// 取消隐藏文件
func unhideFile(path string) {
	ptr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		panic(err)
		return
	}
	attrs, err := syscall.GetFileAttributes(ptr)
	if err != nil {
		panic(err)
		return
	}
	_ = syscall.SetFileAttributes(ptr, attrs&^syscall.FILE_ATTRIBUTE_HIDDEN)
}

// 监听目录变化
func watchDir(dir string, done chan struct{}) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
		return
	}
	defer watcher.Close()

	err = watcher.Add(dir)
	if err != nil {
		panic(err)
		return
	}

	for {
		select {
		case <-done:
			return // 收到停止信号时退出
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
		// 处理错误
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

// 处理自定义目录的菜单
func addCustomDirMenu(dir *CustomDir, parentMenu *systray.MenuItem) {
	menuName := dir.Alias
	if menuName == "" {
		menuName = dir.Path
	}
	dirMenu := parentMenu.AddSubMenuItem(menuName, "")
	mActivate := dirMenu.AddSubMenuItem("激活", "")
	mShow := dirMenu.AddSubMenuItem("显示", "")
	mHide := dirMenu.AddSubMenuItem("隐藏", "")
	//dirMenu.AddSeparator()
	mRemove := dirMenu.AddSubMenuItem("移除", "")

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

// 处理添加新目录
func addNewCustomDir(parentMenu *systray.MenuItem) {
	path, err := zenity.SelectFile(
		zenity.Title("选择目录"),
		zenity.Directory(),
	)
	if err != nil {
		if err == zenity.ErrCanceled {
			return // 用户取消了选择
		}
		zenity.Error("选择目录时出错: " + err.Error())
		return
	}

	alias, err := zenity.Entry(
		"为这个目录输入一个别名（可选）",
		zenity.Title("设置别名"),
		//zenity.AlwaysOnTop(),
	)
	if err != nil {
		if err == zenity.ErrCanceled {
			alias = "" // 用户没有输入别名，使用空字符串
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

// 切换目录的激活状态
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

// 显示自定义目录
func showCustomDir(dir *CustomDir, showItem, hideItem *systray.MenuItem) {
	dir.Hidden = false
	showItem.Check()
	hideItem.Uncheck()
	unhideDotFiles(dir.Path)
	saveConfig()
}

// 隐藏自定义目录
func hideCustomDir(dir *CustomDir, showItem, hideItem *systray.MenuItem) {
	dir.Hidden = true
	hideItem.Check()
	showItem.Uncheck()
	hideDotFiles(dir.Path)
	saveConfig()
}

// 移除自定义目录
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

// 加载配置
func loadConfig() {
	configPath := filepath.Join(os.Getenv("USERPROFILE"), ".hidell", "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		// 如果文件不存在，创建一个空的配置
		if os.IsNotExist(err) {
			customDirs = []CustomDir{}
			return
		}
		// 处理其他错误
		panic(err)
	}

	err = json.Unmarshal(data, &customDirs)
	if err != nil {
		panic(err)
	}
}

// 保存配置
func saveConfig() {
	configPath := filepath.Join(os.Getenv("USERPROFILE"), ".hidell", "config.json")
	data, err := json.MarshalIndent(customDirs, "", "  ")
	if err != nil {
		panic(err)
	}

	err = os.MkdirAll(filepath.Dir(configPath), 0755)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		panic(err)
	}
}

// 添加 startWatching 方法
func (dir *CustomDir) startWatching() {
	if dir.watcher != nil {
		return // 已经在监视中
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

// 添加 stopWatching 方法
func (dir *CustomDir) stopWatching() {
	if dir.watcher == nil {
		return // 没有在监视
	}

	close(dir.done)
	dir.watcher.Close()
	dir.watcher = nil
}

// 添加 watchDir 方法
func (dir *CustomDir) watchDir() {
	for {
		select {
		case <-dir.done:
			return // 收到停止信号时退出
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
