package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigNormalizesPaths(t *testing.T) {
	// Create a tempdir
	tmp := t.TempDir()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(prev) })

	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}

	// Write a config file
	reposJson := `{"repos":[{"path":"repo1","ref":"main","submodules":false}]}`
	if err := os.WriteFile(ConfigFile, []byte(reposJson), 0o644); err != nil {
		t.Fatal(err)
	}

	// Try loading the config
	cfg, err := loadConfig(ConfigFile)
	if err != nil {
		t.Fatalf("loadConfig error: %v", err)
	}
	if len(cfg.Repos) != 1 {
		t.Fatalf("expected 1 repo, got %d", len(cfg.Repos))
	}
	p := cfg.Repos[0].Path
	if !filepath.IsAbs(p) {
		t.Fatalf("expected absolute path, got %s", p)
	}
}
