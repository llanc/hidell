package main

import (
	_ "embed"
	"github.com/fsnotify/fsnotify"
	"github.com/getlantern/systray"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
	"github.com/skratchdot/open-golang/open"
	"golang.org/x/sys/windows/registry"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

//go:embed hidell.ico
var logo []byte

var mw *walk.MainWindow
var systrayToolTip = "HIDE Like Linux"

// 控制 watchDir goroutine 的通道
var watcherDone chan struct{}

func main() {
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
	mAutoStart := systray.AddMenuItem("开机自启", "设置程序开机自启")
	if isAutoStartEnabled() {
		mAutoStart.Check()
	}
	systray.AddSeparator()
	mAbout := systray.AddMenuItem("关于HIDELL V1.1", "")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("退出", "")

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

func runHomeWindow() {
	var err error
	mw, err = walk.NewMainWindow()
	if err != nil {
		panic(err)
	}

	err = MainWindow{
		AssignTo: &mw,
		Title:    "HIDELL",
		MinSize:  Size{Width: 300, Height: 400},
		Size:     Size{Width: 300, Height: 400},
		Layout:   VBox{},
		Children: []Widget{
			HSplitter{
				Children: []Widget{
					PushButton{Text: "激活"},
					PushButton{Text: "隐藏"},
					PushButton{Text: "显示"},
					PushButton{Text: "添加目录"},
				},
			},
		},
	}.Create()
	if err != nil {
		panic(err)
	}

	// 移除右上角的最大化按钮
	style := win.GetWindowLong(mw.Handle(), win.GWL_STYLE)
	style &^= win.WS_MAXIMIZEBOX
	win.SetWindowLong(mw.Handle(), win.GWL_STYLE, style)

	// 处理最小化事件
	mw.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		if reason == walk.CloseReasonUnknown {
			*canceled = true
			mw.Hide()
		}
	})

	mw.Run()
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
