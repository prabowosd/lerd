package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/geodro/lerd/internal/config"
	"github.com/geodro/lerd/internal/sitedoctor"
)

func TestFindSite_ByName(t *testing.T) {
	sites := []config.Site{{Name: "alpha"}, {Name: "beta", Path: "/p/beta"}}
	got, err := findSite(sites, "beta")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Path != "/p/beta" {
		t.Errorf("findSite returned the wrong site: %+v", got)
	}
}

func TestFindSite_NotFoundListsKnown(t *testing.T) {
	sites := []config.Site{{Name: "alpha"}, {Name: "beta"}}
	_, err := findSite(sites, "missing")
	if err == nil {
		t.Fatal("expected an error for an unknown site")
	}
	for _, want := range []string{"missing", "alpha", "beta"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q should mention %q", err.Error(), want)
		}
	}
}

func TestRenderSiteDoctor_ShowsChecksAndSummary(t *testing.T) {
	site := config.Site{Name: "shop", Domains: []string{"shop.test"}, Path: "/www/shop"}
	resp := sitedoctor.Response{
		Checks: []sitedoctor.Check{
			{Name: "app_key", Status: sitedoctor.StatusOK},
			{Name: "env_drift", Status: sitedoctor.StatusWarn, Detail: "FOO missing"},
			{Name: "migrations", Status: sitedoctor.StatusFail, Detail: "2 pending", Fix: "migrate"},
		},
		Failures: 1,
		Warnings: 1,
	}
	var buf bytes.Buffer
	renderSiteDoctor(&buf, false, site, resp)
	out := buf.String()

	for _, want := range []string{
		"shop (shop.test)", "/www/shop",
		"app_key", "env_drift", "FOO missing",
		"migrations", "2 pending", "migrate",
		"1 failure(s), 1 warning(s) found.",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("rendered output missing %q:\n%s", want, out)
		}
	}
}

func TestRenderSiteDoctor_AllPassed(t *testing.T) {
	site := config.Site{Name: "shop", Path: "/www/shop"}
	resp := sitedoctor.Response{Checks: []sitedoctor.Check{{Name: "app_key", Status: sitedoctor.StatusOK}}}
	var buf bytes.Buffer
	renderSiteDoctor(&buf, false, site, resp)
	if !strings.Contains(buf.String(), "All app checks passed.") {
		t.Errorf("expected an all-passed summary, got:\n%s", buf.String())
	}
}

func TestRenderSiteDoctor_NoApplicableChecks(t *testing.T) {
	site := config.Site{Name: "static", Path: "/www/static"}
	var buf bytes.Buffer
	renderSiteDoctor(&buf, false, site, sitedoctor.Response{})
	if !strings.Contains(buf.String(), "No app-level checks apply") {
		t.Errorf("expected the no-checks note, got:\n%s", buf.String())
	}
}
