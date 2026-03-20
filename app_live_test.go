package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLiveImportByAPIKey(t *testing.T) {
	apiKey := os.Getenv("WS_LIVE_API_KEY")
	if apiKey == "" {
		t.Skip("WS_LIVE_API_KEY is not set")
	}

	tmp := t.TempDir()
	origAppData := os.Getenv("APPDATA")
	origLocalAppData := os.Getenv("LOCALAPPDATA")
	if err := os.Setenv("APPDATA", tmp); err != nil {
		t.Fatalf("set APPDATA: %v", err)
	}
	if err := os.Setenv("LOCALAPPDATA", filepath.Join(tmp, "Local")); err != nil {
		t.Fatalf("set LOCALAPPDATA: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Setenv("APPDATA", origAppData)
		_ = os.Setenv("LOCALAPPDATA", origLocalAppData)
	})

	app := NewApp()
	app.startup(context.Background())

	results := app.ImportByAPIKey([]APIKeyItem{{APIKey: apiKey, Remark: "live"}})
	if len(results) != 1 {
		t.Fatalf("ImportByAPIKey() results len = %d, want 1", len(results))
	}
	if !results[0].Success {
		t.Fatalf("ImportByAPIKey() success = false, error = %q", results[0].Error)
	}

	accounts := app.GetAllAccounts()
	if len(accounts) != 1 {
		t.Fatalf("accounts len = %d, want 1", len(accounts))
	}
	acc := accounts[0]
	if acc.PlanName == "" || acc.PlanName == "unknown" {
		t.Fatalf("PlanName = %q", acc.PlanName)
	}
	if acc.DailyRemaining == "" || acc.WeeklyRemaining == "" {
		t.Fatalf("remaining = %q / %q", acc.DailyRemaining, acc.WeeklyRemaining)
	}
	if acc.Email == "" {
		t.Fatal("Email should not be empty")
	}
}
