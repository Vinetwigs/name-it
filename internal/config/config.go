// Package config defines the user-tunable settings for a session. Values are
// kept in memory only; Name It has no persistent state to corrupt.
package config

type Config struct {
	Aggressiveness       string `json:"aggressiveness"`
	NotificationsEnabled bool   `json:"notificationsEnabled"`
	NamingMode           string `json:"namingMode"`
	CustomPattern        string `json:"customPattern"`
}

func Default() Config {
	return Config{
		Aggressiveness:       "annoying",
		NotificationsEnabled: true,
		NamingMode:           "every_file",
		CustomPattern:        "data_filename_vX.extension",
	}
}
