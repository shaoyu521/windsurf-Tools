package services

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallRootsForCandidate_FilePathUsesParentDir(t *testing.T) {
	root := t.TempDir()
	exe := filepath.Join(root, "bin", "windsurf")
	if err := os.MkdirAll(filepath.Dir(exe), 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(exe, []byte("#!/bin/sh"), 0755); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	got := installRootsForCandidate(exe)
	want := []string{filepath.Join(root, "bin"), root}
	if len(got) != len(want) {
		t.Fatalf("installRootsForCandidate len = %d, want %d (%#v)", len(got), len(want), got)
	}
	for i := range want {
		if filepath.Clean(got[i]) != filepath.Clean(want[i]) {
			t.Fatalf("installRootsForCandidate[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestResolveLinuxWindsurfExecutablePathPrefersInstallRoot(t *testing.T) {
	originalLookPath := lookPathFn
	originalEval := evalSymlinksFn
	lookPathFn = func(string) (string, error) { return "", os.ErrNotExist }
	evalSymlinksFn = func(path string) (string, error) { return path, nil }
	defer func() {
		lookPathFn = originalLookPath
		evalSymlinksFn = originalEval
	}()

	root := t.TempDir()
	exe := filepath.Join(root, "windsurf")
	if err := os.WriteFile(exe, []byte("#!/bin/sh"), 0755); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	got := resolveLinuxWindsurfExecutablePath(root)
	if filepath.Clean(got) != filepath.Clean(exe) {
		t.Fatalf("resolveLinuxWindsurfExecutablePath() = %q, want %q", got, exe)
	}
}

func TestLinuxWindsurfInstallCandidatesIncludesLookPathResult(t *testing.T) {
	originalLookPath := lookPathFn
	originalEval := evalSymlinksFn
	lookPathFn = func(string) (string, error) { return "/usr/bin/windsurf", nil }
	evalSymlinksFn = func(path string) (string, error) { return "/opt/Windsurf/windsurf", nil }
	defer func() {
		lookPathFn = originalLookPath
		evalSymlinksFn = originalEval
	}()

	got := linuxWindsurfInstallCandidates("/home/tester")
	want := []string{"/usr/bin/windsurf", "/opt/Windsurf/windsurf"}
	for _, candidate := range want {
		found := false
		for _, path := range got {
			if filepath.Clean(path) == filepath.Clean(candidate) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("linux install candidates missing %q in %#v", candidate, got)
		}
	}
}
