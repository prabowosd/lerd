package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetTinkerForDir_LaravelBuiltin(t *testing.T) {
	setDataDir(t)
	setConfigDir(t)

	site := t.TempDir()
	if err := os.WriteFile(filepath.Join(site, "artisan"), []byte("#!/usr/bin/env php\n"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(site, "vendor", "laravel", "tinker"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := AddSite(Site{Name: "myapp", Domains: []string{"myapp.test"}, Path: site, Framework: "laravel"}); err != nil {
		t.Fatal(err)
	}

	got := GetTinkerForDir(site)
	if got == nil {
		t.Fatal("expected tinker spec for Laravel site")
	}
	if len(got.Command) == 0 || got.Command[0] != "artisan" {
		t.Errorf("expected first command to be 'artisan', got %v", got.Command)
	}
	if got.ExecuteFlag != "--execute" {
		t.Errorf("ExecuteFlag = %q, want --execute", got.ExecuteFlag)
	}
}

func TestGetTinkerForDir_LaravelMissingTinkerPackage(t *testing.T) {
	setDataDir(t)
	setConfigDir(t)

	site := t.TempDir()
	if err := os.WriteFile(filepath.Join(site, "artisan"), []byte("#!/usr/bin/env php\n"), 0755); err != nil {
		t.Fatal(err)
	}
	// no vendor/laravel/tinker on purpose
	if err := AddSite(Site{Name: "myapp", Domains: []string{"myapp.test"}, Path: site, Framework: "laravel"}); err != nil {
		t.Fatal(err)
	}

	if got := GetTinkerForDir(site); got != nil {
		t.Errorf("expected nil (requires_package not satisfied), got %+v", got)
	}
}

func TestGetTinkerForDir_NoSite(t *testing.T) {
	setDataDir(t)
	setConfigDir(t)

	if got := GetTinkerForDir(t.TempDir()); got != nil {
		t.Errorf("expected nil for unregistered path, got %+v", got)
	}
}

func TestGetTinkerForDir_NoFrameworkAssigned(t *testing.T) {
	setDataDir(t)
	setConfigDir(t)

	site := t.TempDir()
	if err := AddSite(Site{Name: "myapp", Domains: []string{"myapp.test"}, Path: site}); err != nil {
		t.Fatal(err)
	}

	if got := GetTinkerForDir(site); got != nil {
		t.Errorf("expected nil for site with no framework, got %+v", got)
	}
}
