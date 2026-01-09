package config

import "CloudStorageProject-FileServer/pkg/tools"

type Config struct {
	Port      int    `yaml:"Port"`
	IPAddress string `yaml:"IPAddress"`
	FilesDir  string `yaml:"FilesDir"`
}

func ReadConfig() (*Config, error) {
	port := tools.GetEnvAsInt("SERVER_PORT", 11682)
	ip := tools.GetEnv("SERVER_IP", "127.0.0.1")
	dir := tools.GetEnv("SERVER_FILE_DIR", "C:/Files")
	config := &Config{
		Port:      port,
		IPAddress: ip,
		FilesDir:  dir,
	}
	return config, nil
}
