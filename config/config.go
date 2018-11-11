package config

import (
	"os"
)

type Config struct {
	Port      string
	SharedKey string
}

func NewConfig() Config {
	return Config{
		Port:      get("PORT", "8081"),
		SharedKey: get("POCAMO_KEYRT", "0x24FEEDFACEDEADBEEFCAFE"),
	}
}

func get(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}
