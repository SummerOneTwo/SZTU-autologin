package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.ISP != "cucc" {
		t.Errorf("Expected ISP cucc, got %s", cfg.ISP)
	}
	if cfg.Area != "dormitory" {
		t.Errorf("Expected Area dormitory, got %s", cfg.Area)
	}
	if cfg.ACID != "17" {
		t.Errorf("Expected ACID 17, got %s", cfg.ACID)
	}
	if !cfg.AutoReconnect {
		t.Error("Expected AutoReconnect true")
	}
	if cfg.CheckInterval != 300 {
		t.Errorf("Expected CheckInterval 300, got %d", cfg.CheckInterval)
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name:    "valid config",
			cfg:     Config{Username: "test", Password: "pass"},
			wantErr: false,
		},
		{
			name:    "empty username",
			cfg:     Config{Username: "", Password: "pass"},
			wantErr: true,
		},
		{
			name:    "empty password",
			cfg:     Config{Username: "test", Password: ""},
			wantErr: true,
		},
		{
			name:    "both empty",
			cfg:     Config{Username: "", Password: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigGetFullUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		isp      string
		want     string
	}{
		{
			name:     "cucc",
			username: "test",
			isp:      "cucc",
			want:     "test@cucc",
		},
		{
			name:     "cmcc",
			username: "test",
			isp:      "cmcc",
			want:     "test@cmcc",
		},
		{
			name:     "chinanet",
			username: "test",
			isp:      "chinanet",
			want:     "test@chinanet",
		},
		{
			name:     "unknown isp",
			username: "test",
			isp:      "unknown",
			want:     "test@cucc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{Username: tt.username, ISP: tt.isp}
			if got := cfg.GetFullUsername(); got != tt.want {
				t.Errorf("GetFullUsername() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	_, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}

	testExe := filepath.Join(tmpDir, "test.exe")
	if err := os.WriteFile(testExe, []byte("test"), 0755); err != nil {
		t.Fatal(err)
	}

	oldOsExecutable := osExecutable
	osExecutable = func() (string, error) {
		return testExe, nil
	}
	defer func() { osExecutable = oldOsExecutable }()

	testCfg := Config{
		Username:      "testuser",
		Password:      "testpass",
		ISP:           "cmcc",
		Area:          "teaching",
		ACID:          "18",
		AutoReconnect: false,
		CheckInterval: 600,
	}

	if err := SaveConfig(testCfg); err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	loadedCfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if loadedCfg.Username != testCfg.Username {
		t.Errorf("Username mismatch: got %v, want %v", loadedCfg.Username, testCfg.Username)
	}
	if loadedCfg.Password != testCfg.Password {
		t.Errorf("Password mismatch")
	}
	if loadedCfg.ISP != testCfg.ISP {
		t.Errorf("ISP mismatch: got %v, want %v", loadedCfg.ISP, testCfg.ISP)
	}
}

var osExecutable = os.Executable
