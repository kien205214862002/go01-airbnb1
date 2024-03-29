package config

import (
	"errors"

	"github.com/spf13/viper"
)

type Config struct {
	App   AppConfig
	MySQL MySQLConfig
}

type AppConfig struct {
	Version string
	Port    string
	Mode    string
	Secret  string
}

type MySQLConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func LoadConfig(filename string) (*Config, error) {
	v := viper.New()

	// Cấu hình để viper biết đọc config từ đâu
	// v.AddConfigPath("config")
	// v.SetConfigName("config-local")
	// v.SetConfigType("yml")

	v.SetConfigFile(filename)
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		// Kiểm tra có phải là lỗi không tìm thấy file config hay không
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, errors.New("config file not found")
		}
		return nil, err
	}

	var c Config
	if err := v.Unmarshal(&c); err != nil {
		return nil, err
	}

	return &c, nil
}
