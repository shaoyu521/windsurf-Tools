package services

import (
	"errors"
	"testing"
)

func TestBuildPrivilegeCommandForLinuxPrefersPkexec(t *testing.T) {
	lookups := map[string]string{
		"install": "/usr/bin/install",
		"pkexec":  "/usr/bin/pkexec",
		"sudo":    "/usr/bin/sudo",
	}
	lookPath := func(name string) (string, error) {
		if path, ok := lookups[name]; ok {
			return path, nil
		}
		return "", errors.New("not found")
	}

	name, args, err := buildPrivilegeCommand("linux", 1000, lookPath, "install", "-m", "0644", "/tmp/src", "/etc/hosts")
	if err != nil {
		t.Fatalf("buildPrivilegeCommand() error = %v", err)
	}
	if name != "/usr/bin/pkexec" {
		t.Fatalf("name = %q, want %q", name, "/usr/bin/pkexec")
	}
	if len(args) < 1 || args[0] != "/usr/bin/install" {
		t.Fatalf("args[0] = %q, want /usr/bin/install (%#v)", firstArg(args), args)
	}
}

func TestBuildPrivilegeCommandForLinuxFallsBackToSudo(t *testing.T) {
	lookups := map[string]string{
		"install": "/usr/bin/install",
		"sudo":    "/usr/bin/sudo",
	}
	lookPath := func(name string) (string, error) {
		if path, ok := lookups[name]; ok {
			return path, nil
		}
		return "", errors.New("not found")
	}

	name, args, err := buildPrivilegeCommand("linux", 1000, lookPath, "install", "/tmp/src", "/etc/hosts")
	if err != nil {
		t.Fatalf("buildPrivilegeCommand() error = %v", err)
	}
	if name != "/usr/bin/sudo" {
		t.Fatalf("name = %q, want %q", name, "/usr/bin/sudo")
	}
	if len(args) < 1 || args[0] != "/usr/bin/install" {
		t.Fatalf("args[0] = %q, want /usr/bin/install (%#v)", firstArg(args), args)
	}
}

func TestBuildPrivilegeCommandForLinuxRootRunsDirectly(t *testing.T) {
	lookPath := func(name string) (string, error) {
		return "/usr/bin/" + name, nil
	}

	name, args, err := buildPrivilegeCommand("linux", 0, lookPath, "install", "a", "b")
	if err != nil {
		t.Fatalf("buildPrivilegeCommand() error = %v", err)
	}
	if name != "/usr/bin/install" {
		t.Fatalf("name = %q, want %q", name, "/usr/bin/install")
	}
	if len(args) != 2 || args[0] != "a" || args[1] != "b" {
		t.Fatalf("args = %#v, want direct args", args)
	}
}

func TestBuildPrivilegeCommandForLinuxRequiresEscalationTool(t *testing.T) {
	lookups := map[string]string{
		"install": "/usr/bin/install",
	}
	lookPath := func(name string) (string, error) {
		if path, ok := lookups[name]; ok {
			return path, nil
		}
		return "", errors.New("not found")
	}

	_, _, err := buildPrivilegeCommand("linux", 1000, lookPath, "install", "/tmp/src", "/etc/hosts")
	if err == nil {
		t.Fatal("buildPrivilegeCommand() expected error when pkexec/sudo missing")
	}
}

func firstArg(args []string) string {
	if len(args) == 0 {
		return ""
	}
	return args[0]
}
