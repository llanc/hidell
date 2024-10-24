package configs

type Config struct {
	CustomDirs []CustomDirs `json:"customDirs"`
	Language   string       `json:"language"`
}

type CustomDirs struct {
	Path   string `json:"path"`
	Alias  string `json:"alias"`
	Active bool   `json:"active"`
	Hidden bool   `json:"hidden"`
}
