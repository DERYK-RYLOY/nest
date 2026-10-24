package config

import (
	"os"
	"path/filepath"
	"testing"
)

// isolate points %APPDATA% at a temp folder so tests never touch the real
// NEST config.
func isolate(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("APPDATA", dir)
	return dir
}

func TestSaveLoadRoundTrip(t *testing.T) {
	isolate(t)
	base := t.TempDir()
	a := filepath.Join(base, "A")
	b := filepath.Join(base, "B")

	cfg := Default()
	cfg.Sources = []string{a, b, a}
	cfg.CustomRules = []CustomRule{
		{Ext: "Blend", Target: "_Blender"},
		{Ext: ".zip", Target: "Pictures"},
		{Ext: ".bad", Target: "///"},
	}
	cfg.InstantDelay = 5
	cfg.MinAgeSeconds = -4

	if err := cfg.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}
	got := Load()

	if len(got.Sources) != 2 {
		t.Errorf("want duplicate sources removed, got %v", got.Sources)
	}
	if len(got.CustomRules) != 2 {
		t.Fatalf("want invalid rule dropped, got %+v", got.CustomRules)
	}
	if got.CustomRules[0].Ext != ".blend" || got.CustomRules[0].Target != "_Blender" {
		t.Errorf("extension should be lowercased: %+v", got.CustomRules[0])
	}
	if got.CustomRules[1].Target != "pictures" {
		t.Errorf("known-folder target should be canonical: %+v", got.CustomRules[1])
	}
	if got.InstantDelay != 10 {
		t.Errorf("instant delay: want 10, got %d", got.InstantDelay)
	}
	if got.MinAgeSeconds != 60 {
		t.Errorf("min age: want 60 for a negative value, got %d", got.MinAgeSeconds)
	}
}

func TestEmptySourceListIsRespected(t *testing.T) {
	isolate(t)
	cfg := Default()
	cfg.Sources = nil
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}
	if got := Load(); len(got.Sources) != 0 {
		t.Errorf("deliberately empty source list was refilled with %v", got.Sources)
	}
}

func TestSaveLeavesNoTempFiles(t *testing.T) {
	dir := isolate(t)
	cfg := Default()
	if err := cfg.Save(); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(filepath.Join(dir, "Nest"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "config.json" {
		t.Errorf("unexpected files after save: %v", entries)
	}
}

func TestSanitizeTarget(t *testing.T) {
	cases := map[string]string{
		"":              "",
		".":             "",
		"..":            "",
		"_":             "",
		"  documents  ": "documents",
		"Documents":     "documents",
		"_Blender":      "_Blender",
		"a/b\\c":        "abc",
		"_My:Stuff?":    "_MyStuff",
	}
	for in, want := range cases {
		if got := SanitizeTarget(in); got != want {
			t.Errorf("SanitizeTarget(%q) = %q, want %q", in, got, want)
		}
	}
}
