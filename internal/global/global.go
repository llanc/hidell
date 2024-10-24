package global

import (
	"github.com/fsnotify/fsnotify"
	"hidell/configs"
)

var Watcher, _ = fsnotify.NewWatcher()
var Conf configs.Config
var Translations map[string]map[string]string
