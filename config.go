package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func LoadConfig() {
	configPath := filepath.Join(getUserHome(), ".hidell", "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			customDirs = []CustomDir{}
			return
		}
		panic(err)
	}

	err = json.Unmarshal(data, &customDirs)
	if err != nil {
		panic(err)
	}
}

func saveConfig() {
	configPath := filepath.Join(getUserHome(), ".hidell", "config.json")
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
