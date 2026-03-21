//go:build !windows

package main

// startTray 在非 Windows 平台禁用，避免托盘依赖阻塞 macOS 发布构建。
func (a *App) startTray() {}

func traySupported() bool { return false }
