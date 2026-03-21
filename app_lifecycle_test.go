package main

import (
	"context"
	"testing"
	"time"

	"github.com/wailsapp/wails/v2/pkg/options"

	"windsurf-tools-wails/backend/store"
)

func TestShutdownCleansMitmEnvironment(t *testing.T) {
	app := NewApp()
	called := 0
	app.cleanupMitmOnExitFn = func() error {
		called++
		return nil
	}

	app.shutdown(context.Background())

	if called != 1 {
		t.Fatalf("shutdown() cleanup calls = %d, want 1", called)
	}
}

func TestActivateExistingWindowUsesToolbarModeWhenEnabled(t *testing.T) {
	s, err := store.NewStoreInPaths(t.TempDir())
	if err != nil {
		t.Fatalf("NewStoreInPaths() error = %v", err)
	}
	settings := s.GetSettings()
	settings.ShowDesktopToolbar = true
	if err := s.UpdateSettings(settings); err != nil {
		t.Fatalf("UpdateSettings() error = %v", err)
	}

	app := NewApp()
	app.store = s

	var (
		called      bool
		showToolbar bool
	)
	app.activateExistingAppFn = func(v bool) {
		called = true
		showToolbar = v
	}

	app.activateExistingWindow()

	if !called {
		t.Fatal("activateExistingWindow() did not trigger activation hook")
	}
	if !showToolbar {
		t.Fatal("activateExistingWindow() should restore toolbar layout")
	}
}

func TestActivateExistingWindowUsesMainWindowWhenToolbarDisabled(t *testing.T) {
	s, err := store.NewStoreInPaths(t.TempDir())
	if err != nil {
		t.Fatalf("NewStoreInPaths() error = %v", err)
	}

	app := NewApp()
	app.store = s

	called := make(chan bool, 1)
	app.activateExistingAppFn = func(v bool) {
		called <- v
	}

	app.onSecondInstanceLaunch(options.SecondInstanceData{})

	select {
	case showToolbar := <-called:
		if showToolbar {
			t.Fatal("onSecondInstanceLaunch() should restore main window when toolbar is disabled")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("onSecondInstanceLaunch() did not trigger activation hook")
	}
}

func TestShouldStartHiddenRequiresTrayWhenToolbarDisabled(t *testing.T) {
	app := NewApp()
	app.silentFromFlag = true
	app.traySupportedFn = func() bool { return false }

	if app.shouldStartHidden() {
		t.Fatal("shouldStartHidden() should ignore silent start when tray is unavailable and toolbar is disabled")
	}
}

func TestShouldStartHiddenAllowsToolbarModeWithoutTray(t *testing.T) {
	s, err := store.NewStoreInPaths(t.TempDir())
	if err != nil {
		t.Fatalf("NewStoreInPaths() error = %v", err)
	}
	settings := s.GetSettings()
	settings.SilentStart = true
	settings.ShowDesktopToolbar = true
	if err := s.UpdateSettings(settings); err != nil {
		t.Fatalf("UpdateSettings() error = %v", err)
	}

	app := NewApp()
	app.store = s
	app.traySupportedFn = func() bool { return false }

	if !app.shouldStartHidden() {
		t.Fatal("shouldStartHidden() should still allow toolbar mode without tray support")
	}
}

func TestOnBeforeCloseIgnoresMinimizeToTrayWhenTrayUnavailable(t *testing.T) {
	s, err := store.NewStoreInPaths(t.TempDir())
	if err != nil {
		t.Fatalf("NewStoreInPaths() error = %v", err)
	}
	settings := s.GetSettings()
	settings.MinimizeToTray = true
	if err := s.UpdateSettings(settings); err != nil {
		t.Fatalf("UpdateSettings() error = %v", err)
	}

	app := NewApp()
	app.store = s
	app.traySupportedFn = func() bool { return false }

	if app.onBeforeClose(context.Background()) {
		t.Fatal("onBeforeClose() should not hide window when tray is unavailable")
	}
}
