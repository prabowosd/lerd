package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/geodro/lerd/internal/config"
)

// isolateLerdHome points both XDG roots at temp dirs so the command's
// LoadCustomService / MaterializeServiceTuning never touch the real home.
func isolateLerdHome(t *testing.T) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", t.TempDir())
}

func runServiceConfig(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := newServiceConfigCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestServiceConfig_PathFlagSeedsAndPrints(t *testing.T) {
	isolateLerdHome(t)
	if err := config.SaveCustomService(&config.CustomService{
		Name:   "mariadb-10-11",
		Image:  "docker.io/library/mariadb:10.11",
		Family: "mariadb",
	}); err != nil {
		t.Fatalf("SaveCustomService: %v", err)
	}

	out, err := runServiceConfig(t, "mariadb-10-11", "--path")
	if err != nil {
		t.Fatalf("config --path: %v", err)
	}
	want := config.ServiceTuningFile("mariadb-10-11")
	if strings.TrimSpace(out) != want {
		t.Errorf("printed path = %q, want %q", strings.TrimSpace(out), want)
	}
	if _, err := os.Stat(want); err != nil {
		t.Errorf("tuning file should be materialized by --path, stat err = %v", err)
	}
}

func TestServiceConfig_RejectsUntunedFamily(t *testing.T) {
	isolateLerdHome(t)
	if err := config.SaveCustomService(&config.CustomService{
		Name:   "redis",
		Image:  "docker.io/library/redis:7",
		Family: "redis",
	}); err != nil {
		t.Fatalf("SaveCustomService: %v", err)
	}

	_, err := runServiceConfig(t, "redis", "--path")
	if err == nil || !strings.Contains(err.Error(), "does not support tuning") {
		t.Errorf("expected untuned-family error, got: %v", err)
	}
}

func TestServiceConfig_RejectsUninstalledService(t *testing.T) {
	isolateLerdHome(t)
	_, err := runServiceConfig(t, "ghost", "--path")
	if err == nil || !strings.Contains(err.Error(), "not installed") {
		t.Errorf("expected not-installed error, got: %v", err)
	}
}
