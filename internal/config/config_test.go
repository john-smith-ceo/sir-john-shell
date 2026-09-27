package config

import (
	"testing"
)

func TestSaveLoadProfile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("SJS_USER", "")

	want := UserConfig{
		User:       "Sir John Smith",
		FontFamily: "ui-monospace",
		Theme:      map[string]any{"scheme": "dracula"},
		UI:         map[string]any{"scanlines": false},
	}
	if err := Save(want); err != nil {
		t.Fatalf("save profile: %v", err)
	}
	got, err := Load()
	if err != nil {
		t.Fatalf("load profile: %v", err)
	}
	if got.User != want.User || got.FontFamily != want.FontFamily {
		t.Fatalf("profile identity mismatch: got %#v", got)
	}
	if got.Theme["scheme"] != "dracula" || got.UI["scanlines"] != false {
		t.Fatalf("profile preferences mismatch: got %#v", got)
	}
}
