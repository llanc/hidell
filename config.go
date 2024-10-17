package main

import (
	"encoding/json"
	"golang.org/x/sys/windows"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

type Config struct {
	CustomDirs []CustomDir `json:"customDirs"`
	Language   string      `json:"language"`
}

var config Config

func LoadConfig() {
	configPath := filepath.Join(getUserHome(), ".hidell", "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			config = Config{
				CustomDirs: []CustomDir{},
				Language:   getDefaultLanguage(),
			}
			saveConfig()
			return
		}
		panic(err)
	}

	// 首先尝试解析完整的新配置
	err = json.Unmarshal(data, &config)
	if err == nil {
		// 成功解析新配置
		if config.Language == "" {
			config.Language = getDefaultLanguage()
			saveConfig()
		}
	} else {
		// 如果解析失败，尝试解析旧版本的配置文件
		var oldConfig []CustomDir
		err = json.Unmarshal(data, &oldConfig)
		if err != nil {
			panic(err)
		}
		// 使用旧配置并创建新的配置结构
		config = Config{
			CustomDirs: oldConfig,
			Language:   getDefaultLanguage(),
		}
		// 保存更新后的配置
		saveConfig()
	}

	customDirs = config.CustomDirs
}

func saveConfig() {
	configPath := filepath.Join(getUserHome(), ".hidell", "config.json")
	data, err := json.MarshalIndent(config, "", "  ")
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

func getDefaultLanguage() string {
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
