package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	ISP           string `json:"isp"`
	ACID          string `json:"ac_id"`
	AutoReconnect bool   `json:"auto_reconnect"`
	CheckInterval int    `json:"check_interval"`
}

func DefaultConfig() Config {
	return Config{
		ISP:           "cucc",
		ACID:          "17",
		AutoReconnect: false,
		CheckInterval: 300,
	}
}

func getConfigPath() string {
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		return filepath.Join(dir, "config.json")
	}
	return "config.json"
}

func LoadConfig() (Config, error) {
	path := getConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return Config{}, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func SaveConfig(cfg Config) error {
	path := getConfigPath()
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func (c Config) GetFullUsername() string {
	suffixes := map[string]string{
		"cucc":     "@cucc",
		"cmcc":     "@cmcc",
		"chinanet": "@chinanet",
	}
	suffix := suffixes[c.ISP]
	if suffix == "" {
		suffix = "@cucc"
	}
	return c.Username + suffix
}

func (c Config) Validate() error {
	if c.Username == "" {
		return fmt.Errorf("用户名不能为空")
	}
	if c.Password == "" {
		return fmt.Errorf("密码不能为空")
	}
	return nil
}
