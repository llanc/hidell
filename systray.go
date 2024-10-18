package main

import (
	"github.com/getlantern/systray"
	"github.com/skratchdot/open-golang/open"
	"os"
	"os/exec"
)

var systrayToolTip = "HIDE Like Linux"

// 全局变量来存储菜单项引用
var (
	mUserHomeActivate *systray.MenuItem
	mUserHomeHide     *systray.MenuItem
	mUserHomeShow     *systray.MenuItem
	mCustom           *systray.MenuItem
	mSettings         *systray.MenuItem
	mEnglish          *systray.MenuItem
	mChinese          *systray.MenuItem
	mAddDir           *systray.MenuItem
	mQuit             *systray.MenuItem
	mLanguage         *systray.MenuItem
	mAutoStart        *systray.MenuItem
	mAbout            *systray.MenuItem
)

func systrayInit() {
	systray.SetIcon(logo)
	systray.SetTitle("HIDELL")
	systray.SetTooltip(systrayToolTip)

	mUserHomeActivate = systray.AddMenuItem(t("activate"), t("auto_hide_new_dot_files"))
	mUserHomeActivate.Check()
	systray.AddSeparator()
	mUserHomeHide = systray.AddMenuItem(t("hide"), t("hide_existing_dot_files"))
	mUserHomeShow = systray.AddMenuItem(t("show"), t("show_existing_dot_files"))
	systray.AddSeparator()

	mCustom = systray.AddMenuItem(t("custom"), t("add_custom_directory"))

	systray.AddSeparator()
	mSettings = systray.AddMenuItem(t("settings"), "")
	mLanguage = mSettings.AddSubMenuItem(t("language"), "")
	mEnglish = mLanguage.AddSubMenuItem("English", "")
	mChinese = mLanguage.AddSubMenuItem("中文", "")
	mAutoStart = mSettings.AddSubMenuItem(t("auto_start"), t("set_auto_start"))
	if isAutoStartEnabled() {
		mAutoStart.Check()
	}
	mAbout = mSettings.AddSubMenuItem(t("about_hidell")+" V1.3", "")

	mQuit = systray.AddMenuItem(t("quit"), "")

	mAddDir = mCustom.AddSubMenuItem(t("add_directory"), t("add_new_directory"))

	// 添加自定义目录菜单
	for i := range config.CustomDirs {
		addCustomDirMenu(&customDirs[i], mCustom)
	}

	userHome := getUserHome()
	go handleMenuClicks(userHome)
}

func handleMenuClicks(userHome string) {
	for {
		select {
		case <-mUserHomeActivate.ClickedCh:
			toggleMonitor(mUserHomeActivate)
		case <-mUserHomeHide.ClickedCh:
			toggleHide(userHome)
		case <-mUserHomeShow.ClickedCh:
			toggleShow(userHome)
		case <-mAutoStart.ClickedCh:
			toggleAutoStart()
		case <-mAbout.ClickedCh:
			open.Run("https://github.com/llanc/hidell")
		case <-mQuit.ClickedCh:
			systray.Quit()
			return
		case <-mAddDir.ClickedCh:
			addNewCustomDir(mAddDir)
		case <-mEnglish.ClickedCh:
			config.Language = "en"
			saveConfig()
			restartProgram()
		case <-mChinese.ClickedCh:
			config.Language = "zh"
			saveConfig()
			restartProgram()
		}
	}
}

func restartProgram() {
	executable, err := os.Executable()
	if err != nil {
		panic(err)
	}

	cmd := exec.Command(executable, os.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err = cmd.Start()
	if err != nil {
		panic(err)
	}

	systray.Quit()
	os.Exit(0)
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

func toggleHide(userHome string) {
	mUserHomeHide.Check()
	mUserHomeShow.Uncheck()
	hideDotFiles(userHome)
}

func toggleShow(userHome string) {
	mUserHomeShow.Check()
	mUserHomeHide.Uncheck()
	unhideDotFiles(userHome)
}

func toggleAutoStart() {
	if mAutoStart.Checked() {
		disableAutoStart()
		mAutoStart.Uncheck()
	} else {
		enableAutoStart()
		mAutoStart.Check()
	}
}
