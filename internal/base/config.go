package base

import (
	"encoding/json"
	"hidell/internal/global"
	"os"
	"path/filepath"
)

func SaveConfig() {
	configPath := filepath.Join(GetUserHome(), ".hidell", "config.json")
	data, err := json.MarshalIndent(global.Conf, "", "  ")
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
