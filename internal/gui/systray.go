package gui

import (
	_ "embed"
	"github.com/getlantern/systray"
	"github.com/skratchdot/open-golang/open"
	"hidell/assets"
	"hidell/internal/base"
	"hidell/internal/fileops"
	"hidell/internal/global"
	"hidell/internal/local"
	"hidell/internal/setting"
)

// 全局变量来存储菜单项引用
var (
	mDefaultDirActivate *systray.MenuItem
	mDefaultDirHide     *systray.MenuItem
	mDefaultDirShow     *systray.MenuItem
	mCustom             *systray.MenuItem
	mSettings           *systray.MenuItem
	mEnglish            *systray.MenuItem
	mChinese            *systray.MenuItem
	mAddDir             *systray.MenuItem
	mQuit               *systray.MenuItem
	mLanguage           *systray.MenuItem
	mAutoStart          *systray.MenuItem
	mAbout              *systray.MenuItem
)

var defaultDir = fileops.MakeDirDefault(base.GetUserHome(), "User Home")

func SystrayInit() {
	systray.SetIcon(assets.LogoIcon)
	systray.SetTitle("HIDELL")
	systray.SetTooltip("HIDE Like Linux")

	mDefaultDirActivate = systray.AddMenuItem(local.T("activate"), local.T("auto_hide_new_dot_files"))
	mDefaultDirActivate.Check()
	systray.AddSeparator()
	mDefaultDirHide = systray.AddMenuItem(local.T("hide"), local.T("hide_existing_dot_files"))
	mDefaultDirShow = systray.AddMenuItem(local.T("show"), local.T("show_existing_dot_files"))
	systray.AddSeparator()

	mCustom = systray.AddMenuItem(local.T("custom"), local.T("add_custom_directory"))

	systray.AddSeparator()
	mSettings = systray.AddMenuItem(local.T("settings"), "")
	mLanguage = mSettings.AddSubMenuItem(local.T("language"), "")
	mEnglish = mLanguage.AddSubMenuItem("English", "")
	mChinese = mLanguage.AddSubMenuItem("中文", "")
	mAutoStart = mSettings.AddSubMenuItem(local.T("auto_start"), local.T("set_auto_start"))
	if setting.IsAutoStartEnabled() {
		mAutoStart.Check()
	}
	mAbout = mSettings.AddSubMenuItem(local.T("about_hidell")+" V1.4", "")

	mQuit = systray.AddMenuItem(local.T("quit"), "")

	mAddDir = mCustom.AddSubMenuItem(local.T("add_directory"), local.T("add_new_directory"))

	// 添加自定义目录菜单
	for _, dir := range global.Conf.CustomDirs {
		confDir := fileops.MakeDir(&dir)
		confDir.AddToMenu(mCustom)
	}

	// 默认用户目录激活
	defaultDir.Dir.Active = false
	defaultDir.ToggleActivate(mDefaultDirActivate)

	go handleMenuClicks()

}

func handleMenuClicks() {

	for {
		select {
		case <-mDefaultDirActivate.ClickedCh:
			defaultDir.ToggleActivate(mDefaultDirActivate)
		case <-mDefaultDirHide.ClickedCh:
			defaultDir.Hide(mDefaultDirShow, mDefaultDirHide)
		case <-mDefaultDirShow.ClickedCh:
			defaultDir.Show(mDefaultDirShow, mDefaultDirHide)
		case <-mAutoStart.ClickedCh:
			toggleAutoStart()
		case <-mAbout.ClickedCh:
			open.Run("https://github.com/llanc/hidell")
		case <-mQuit.ClickedCh:
			systray.Quit()
			return
		case <-mAddDir.ClickedCh:
			fileops.AddDir(mCustom)
		case <-mEnglish.ClickedCh:
			global.Conf.Language = "en"
			base.SaveConfig()
			setting.RestartProgram()
		case <-mChinese.ClickedCh:
			global.Conf.Language = "zh"
			base.SaveConfig()
			setting.RestartProgram()
		}
	}
}

func toggleAutoStart() {
	if mAutoStart.Checked() {
		setting.DisableAutoStart()
		mAutoStart.Uncheck()
	} else {
		setting.EnableAutoStart()
		mAutoStart.Check()
	}
}
