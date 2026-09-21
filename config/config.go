package config

import "os"

type Config struct {
	BiosPath string // Path to the GBA BIOS file
}

var config *Config

func NewConfig() *Config {
	if config != nil {
		return config
	}

	config = &Config{}
	biosPath, ok := os.LookupEnv("GBA_BIOS_PATH")
	if !ok {
		config.BiosPath = "bios.bin"
	} else {
		config.BiosPath = biosPath
	}

	return config

}
