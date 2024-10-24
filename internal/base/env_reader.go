package base

import (
	"golang.org/x/sys/windows"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
	"os"
	"syscall"
	"unsafe"
)

func GetUserHome() string {
	return os.Getenv("USERPROFILE")
}

func GetDefaultLanguage() string {
	// 获取用户界面语言
	k32 := windows.NewLazySystemDLL("kernel32.dll")
	getUserDefaultUILanguage := k32.NewProc("GetUserDefaultUILanguage")
	if getUserDefaultUILanguage.Find() == nil {
		ret, _, _ := getUserDefaultUILanguage.Call()
		langID := uint16(ret)
		userLang := windowsLangIDToISO639(langID)
		if userLang != "" {
			return mapLanguage(userLang)
		}
	}

	// 如果无法获取用户界面语言，尝试获取系统区域设置
	systemLang := getSystemLocale()
	return mapLanguage(systemLang)
}

func windowsLangIDToISO639(langID uint16) string {
	// 这里只处理了英语和中文，您可以根据需要添加更多语言
	switch langID {
	case 0x0409: // English (United States)
		return "en"
	case 0x0804: // Chinese (Simplified, PRC)
		return "zh"
	case 0x0C04: // Chinese (Traditional, Hong Kong S.A.R.)
		return "zh"
	case 0x1004: // Chinese (Simplified, Singapore)
		return "zh"
	case 0x0404: // Chinese (Traditional, Taiwan)
		return "zh"
	default:
		return ""
	}
}

func getSystemLocale() string {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getSystemDefaultLocaleName := kernel32.NewProc("GetSystemDefaultLocaleName")

	buf := make([]uint16, 85) // LOCALE_NAME_MAX_LENGTH = 85
	r, _, _ := getSystemDefaultLocaleName.Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if r == 0 {
		return "en" // 默认返回英语
	}

	return syscall.UTF16ToString(buf)
}

func mapLanguage(lang string) string {
	tag := language.Make(lang)
	base, _ := tag.Base()
	langName := display.English.Languages().Name(base)

	switch langName {
	case "English":
		return "en"
	case "Chinese":
		return "zh"
	// 添加其他支持的语言
	default:
		return "en"
	}
}

func GetForegroundWindow() uintptr {
	var user32 = windows.NewLazyDLL("user32.dll")
	hwnd, _, _ := user32.NewProc("GetForegroundWindow").Call()
	return hwnd
}
