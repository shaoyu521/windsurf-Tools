package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"windsurf-tools-wails/backend/models"
)

type Store struct {
	accountsFile string
	settingsFile string
	mu           sync.RWMutex
	accounts     []models.Account
	settings     models.Settings
}

func NewStore() (*Store, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config dir: %w", err)
	}

	appDir := filepath.Join(configDir, "windsurf-tools-wails")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	s := &Store{
		accountsFile: filepath.Join(appDir, "accounts.json"),
		settingsFile: filepath.Join(appDir, "settings.json"),
		accounts:     make([]models.Account, 0),
		settings:     models.DefaultSettings(),
	}

	s.load()
	return s, nil
}

func (s *Store) load() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if b, err := os.ReadFile(s.accountsFile); err == nil {
		_ = json.Unmarshal(b, &s.accounts)
	}
	if b, err := os.ReadFile(s.settingsFile); err == nil {
		_ = json.Unmarshal(b, &s.settings)
	}
}

func (s *Store) saveAccounts() error {
	b, err := json.MarshalIndent(s.accounts, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.accountsFile, b, 0644)
}

func (s *Store) saveSettings() error {
	b, err := json.MarshalIndent(s.settings, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.settingsFile, b, 0644)
}

// ── Account Operations ──

func (s *Store) AddAccount(acc models.Account) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.accounts {
		if AccountsConflict(s.accounts[i], acc) {
			return fmt.Errorf("账号已存在，不可重复导入")
		}
	}
	s.accounts = append(s.accounts, acc)
	return s.saveAccounts()
}

func (s *Store) GetAllAccounts() []models.Account {
	s.mu.RLock()
	defer s.mu.RUnlock()
	copied := make([]models.Account, len(s.accounts))
	copy(copied, s.accounts)
	return copied
}

func (s *Store) GetAccount(id string) (*models.Account, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := range s.accounts {
		if s.accounts[i].ID == id {
			return &s.accounts[i], nil
		}
	}
	return nil, fmt.Errorf("account not found")
}

func (s *Store) UpdateAccount(acc models.Account) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.accounts {
		if s.accounts[i].ID == acc.ID {
			s.accounts[i] = acc
			return s.saveAccounts()
		}
	}
	return fmt.Errorf("account not found")
}

func (s *Store) DeleteAccount(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.accounts {
		if s.accounts[i].ID == id {
			s.accounts = append(s.accounts[:i], s.accounts[i+1:]...)
			return s.saveAccounts()
		}
	}
	return fmt.Errorf("account not found")
}

// ── Settings Operations ──

func (s *Store) GetSettings() models.Settings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.settings
}

func (s *Store) UpdateSettings(st models.Settings) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.settings = st
	return s.saveSettings()
}
