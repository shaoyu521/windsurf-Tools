package services

import (
	"runtime"
	"strings"
)

func currentWindsurfClientPlatform() string {
	return windsurfClientPlatform(runtime.GOOS)
}

func windsurfClientPlatform(goos string) string {
	switch strings.ToLower(strings.TrimSpace(goos)) {
	case "linux":
		return "linux"
	case "darwin":
		return "darwin"
	case "windows":
		return "windows"
	default:
		return "windows"
	}
}
