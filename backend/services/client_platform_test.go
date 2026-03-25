package services

import "testing"

func TestWindsurfClientPlatform(t *testing.T) {
	cases := map[string]string{
		"windows": "windows",
		"linux":   "linux",
		"darwin":  "darwin",
		"WINDOWS": "windows",
		"":        "windows",
		"freebsd": "windows",
	}

	for input, want := range cases {
		if got := windsurfClientPlatform(input); got != want {
			t.Fatalf("windsurfClientPlatform(%q) = %q, want %q", input, got, want)
		}
	}
}
