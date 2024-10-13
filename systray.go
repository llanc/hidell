package main

import (
	"github.com/getlantern/systray"
	"github.com/skratchdot/open-golang/open"
)

var systrayToolTip = "HIDE Like Linux"

func onReady() {
	systray.SetIcon(logo)
	systray.SetTitle("HIDELL")
	systray.SetTooltip(systrayToolTip)

	mMonitor := systray.AddMenuItem("激活", "自动隐藏新生成的点文件/点文件夹")
	mMonitor.Check()
	systray.AddSeparator()
	mHide := systray.AddMenuItem("隐藏", "隐藏已存在的点文件/点文件夹")
	mShow := systray.AddMenuItem("显示", "显示已存在的点文件/点文件夹")
	systray.AddSeparator()

	mOtherDirs := systray.AddMenuItem("其他", "管理其他目录")

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

	userHome := getUserHome()
	go handleMenuClicks(mMonitor, mHide, mShow, mAutoStart, mAbout, mQuit, mAddDir, mOtherDirs, userHome)
}

func handleMenuClicks(mMonitor, mHide, mShow, mAutoStart, mAbout, mQuit, mAddDir *systray.MenuItem, mOtherDirs *systray.MenuItem, userHome string) {
	for {
		select {
		case <-mMonitor.ClickedCh:
			toggleMonitor(mMonitor)
		case <-mHide.ClickedCh:
			toggleHide(mHide, mShow, userHome)
		case <-mShow.ClickedCh:
			toggleShow(mShow, mHide, userHome)
		case <-mAutoStart.ClickedCh:
			toggleAutoStart(mAutoStart)
		case <-mAbout.ClickedCh:
			_ = open.Run("https://www.github.com/llanc/hidell")
		case <-mQuit.ClickedCh:
			systray.Quit()
			return
		case <-mAddDir.ClickedCh:
			addNewCustomDir(mOtherDirs)
		}
	}
}

func toggleMonitor(mMonitor *systray.MenuItem) {
	if mMonitor.Checked() {
		mMonitor.Uncheck()
		if watcherDone != nil {
			close(watcherDone)
			watcherDone = nil
		}
	} else {
		mMonitor.Check()
		watcherDone = make(chan struct{})
		go watchDir(getUserHome(), watcherDone)
	}
}

func toggleHide(mHide, mShow *systray.MenuItem, userHome string) {
	mHide.Check()
	mShow.Uncheck()
	hideDotFiles(userHome)
}

func toggleShow(mShow, mHide *systray.MenuItem, userHome string) {
	mShow.Check()
	mHide.Uncheck()
	unhideDotFiles(userHome)
}

func toggleAutoStart(mAutoStart *systray.MenuItem) {
	if mAutoStart.Checked() {
		disableAutoStart()
		mAutoStart.Uncheck()
	} else {
		enableAutoStart()
		mAutoStart.Check()
	}
}
