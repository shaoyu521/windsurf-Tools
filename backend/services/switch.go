package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// SwitchService handles seamless account switching
type SwitchService struct{}

func NewSwitchService() *SwitchService {
	return &SwitchService{}
}

// WindsurfAuthJSON is the structure of windsurf_auth.json
type WindsurfAuthJSON struct {
	Token     string `json:"token"`
	Email     string `json:"email,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
}

// GetWindsurfAuthPath returns the path to windsurf_auth.json
func (s *SwitchService) GetWindsurfAuthPath() (string, error) {
	var base string

	switch runtime.GOOS {
	case "windows":
		appdata := os.Getenv("APPDATA")
		if appdata == "" {
			home, _ := os.UserHomeDir()
			appdata = filepath.Join(home, "AppData", "Roaming")
		}
		base = filepath.Join(appdata, ".codeium", "windsurf", "config")
	case "darwin":
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, ".codeium", "windsurf", "config")
	default: // linux
		home, _ := os.UserHomeDir()
		base = filepath.Join(home, ".codeium", "windsurf", "config")
	}

	return filepath.Join(base, "windsurf_auth.json"), nil
}

// SwitchAccount writes the token into windsurf_auth.json for seamless switching
func (s *SwitchService) SwitchAccount(token, email string) error {
	authPath, err := s.GetWindsurfAuthPath()
	if err != nil {
		return fmt.Errorf("获取auth路径失败: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(authPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// Backup existing file
	if _, err := os.Stat(authPath); err == nil {
		backupPath := authPath + fmt.Sprintf(".bak.%d", time.Now().Unix())
		if data, err := os.ReadFile(authPath); err == nil {
			_ = os.WriteFile(backupPath, data, 0644)
		}
	}

	// Write new auth
	auth := WindsurfAuthJSON{
		Token:     token,
		Email:     email,
		Timestamp: time.Now().Unix(),
	}

	data, err := json.MarshalIndent(auth, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化auth失败: %w", err)
	}

	if err := os.WriteFile(authPath, data, 0644); err != nil {
		return fmt.Errorf("写入auth文件失败: %w", err)
	}

	return nil
}

// GetCurrentAuth reads the current windsurf_auth.json
func (s *SwitchService) GetCurrentAuth() (*WindsurfAuthJSON, error) {
	authPath, err := s.GetWindsurfAuthPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(authPath)
	if err != nil {
		return nil, fmt.Errorf("读取auth文件失败: %w", err)
	}

	var auth WindsurfAuthJSON
	if err := json.Unmarshal(data, &auth); err != nil {
		return nil, fmt.Errorf("解析auth文件失败: %w", err)
	}
	return &auth, nil
}
