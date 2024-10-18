package main

import (
	"encoding/json"
	"os"
)

var translations map[string]map[string]string

func loadTranslations() {
	translations = make(map[string]map[string]string)

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

		translations[lang] = langTranslations
	}
}

func t(key string) string {
	lang := config.Language
	if _, ok := translations[lang]; !ok {
		lang = "en" // 默认语言
	}

	if translation, ok := translations[lang][key]; ok {
		return translation
	}
	return key
}
