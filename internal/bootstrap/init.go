package bootstrap

import (
	"encoding/json"
	"hidell/configs"
	"hidell/internal/base"
	"hidell/internal/global"
	"os"
	"path/filepath"
)

func Init() {
	config()
	i18n()
}

func config() {
	configPath := filepath.Join(base.GetUserHome(), ".hidell", "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			global.Conf = configs.Config{
				CustomDirs: []configs.CustomDirs{},
				Language:   base.GetDefaultLanguage(),
			}
			base.SaveConfig()
			return
		}
		panic(err)
	}

	// 首先尝试解析完整的新配置
	err = json.Unmarshal(data, &global.Conf)
	if err == nil {
		// 成功解析新配置
		if global.Conf.Language == "" {
			global.Conf.Language = base.GetDefaultLanguage()
			base.SaveConfig()
		}
	} else {
		// 如果解析失败，尝试解析旧版本的配置文件
		var oldConfig []configs.CustomDirs
		err = json.Unmarshal(data, &oldConfig)
		if err != nil {
			panic(err)
		}
		// 使用旧配置并创建新的配置结构
		global.Conf = configs.Config{
			CustomDirs: oldConfig,
			Language:   base.GetDefaultLanguage(),
		}
		// 保存更新后的配置
		base.SaveConfig()
	}
}

func i18n() {
	global.Translations = make(map[string]map[string]string)

	languages := []string{"en", "zh"}
	for _, lang := range languages {
		file, err := os.ReadFile("locales/" + lang + ".json")
		if err != nil {
			panic(err)
		}

		var langTranslations map[string]string
		err = json.Unmarshal(file, &langTranslations)
		if err != nil {
			panic(err)
		}

		global.Translations[lang] = langTranslations
	}
}
