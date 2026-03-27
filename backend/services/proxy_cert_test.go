package services

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLinuxCAInstalledRequiresMatchingSystemCopy(t *testing.T) {
	dir := t.TempDir()
	localCert := filepath.Join(dir, "ca.pem")
	systemCert := filepath.Join(dir, "windsurf-tools-ca.crt")

	if err := os.WriteFile(localCert, []byte("local-cert"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(systemCert, []byte("local-cert"), 0644); err != nil {
		t.Fatal(err)
	}

	if !linuxCAInstalled(localCert, systemCert) {
		t.Fatal("linuxCAInstalled() = false, want true for matching cert copies")
	}
}

func TestLinuxCAInstalledRejectsMissingOrMismatchedSystemCopy(t *testing.T) {
	dir := t.TempDir()
	localCert := filepath.Join(dir, "ca.pem")
	systemCert := filepath.Join(dir, "windsurf-tools-ca.crt")

	if err := os.WriteFile(localCert, []byte("local-cert"), 0644); err != nil {
		t.Fatal(err)
	}
	if linuxCAInstalled(localCert, systemCert) {
		t.Fatal("linuxCAInstalled() = true, want false when system copy is missing")
	}
	if err := os.WriteFile(systemCert, []byte("other-cert"), 0644); err != nil {
		t.Fatal(err)
	}
	if linuxCAInstalled(localCert, systemCert) {
		t.Fatal("linuxCAInstalled() = true, want false for mismatched system copy")
	}
}
