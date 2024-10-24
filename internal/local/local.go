package local

import "hidell/internal/global"

func T(key string) string {
	lang := global.Conf.Language
	if _, ok := global.Translations[lang]; !ok {
		lang = "en" // 默认语言
	}

	if translation, ok := global.Translations[lang][key]; ok {
		return translation
	}
	return key
}
