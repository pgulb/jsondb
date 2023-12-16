package config

import (
	"os"

	"github.com/olebedev/config"
)

func ReadConfig(path string) (string, int, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return "", 0, err
	}
	jsonString := string(file)

	cfg, err := config.ParseJson(jsonString)
	if err != nil {
		return "", 0, err
	}
	jsonsPath, err := cfg.String("JsonsPath")
	if err != nil {
		return "", 0, err
	}
	ReceivingResponseChannelTimeout, err := cfg.Int("ReceivingResponseChannelTimeout")
	if err != nil {
		return "", 0, err
	}
	return jsonsPath, ReceivingResponseChannelTimeout, nil
}
