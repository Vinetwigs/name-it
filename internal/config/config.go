package config

type Config struct {
	Aggressiveness         string   `json:"aggressiveness"`
	DefaultReminderMinutes int      `json:"defaultReminderMinutes"`
	MinimumReminderMinutes int      `json:"minimumReminderMinutes"`
	IgnoredExtensions      []string `json:"ignoredExtensions"`
	IgnoredDirectories     []string `json:"ignoredDirectories"`
	SuspiciousWords        []string `json:"suspiciousWords"`
	NotificationsEnabled   bool     `json:"notificationsEnabled"`
	NamingMode             string   `json:"namingMode"`
	CustomPattern          string   `json:"customPattern"`
}

func Default() Config {
	return Config{Aggressiveness: "annoying", DefaultReminderMinutes: 30, MinimumReminderMinutes: 5, NotificationsEnabled: true,
		IgnoredExtensions: []string{".tmp", ".temp", ".lock"}, IgnoredDirectories: []string{"node_modules", ".git", ".cache"},
		SuspiciousWords: []string{"final", "finale", "def", "definitivo", "definitive", "ultimo", "last", "new", "nuovo", "copy", "copia", "test", "prova", "temp", "temporary", "old", "vecchio", "latest", "vero", "true", "real", "actual"},
		NamingMode:      "every_file",
		CustomPattern:   "data_filename_vX.extension"}
}
