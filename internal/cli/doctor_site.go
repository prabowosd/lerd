package cli

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/geodro/lerd/internal/config"
	"github.com/geodro/lerd/internal/feedback"
	"github.com/geodro/lerd/internal/sitedoctor"
)

// RunSiteDoctorTo resolves the named site and renders its app-level health to w,
// the same checks the dashboard's Laravel Doctor shows but on the command line
// (and so reachable to automation, which the web/TUI panels are not).
func RunSiteDoctorTo(w io.Writer, useColor bool, name string) error {
	reg, err := config.LoadSites()
	if err != nil {
		return err
	}
	site, err := findSite(reg.Sites, name)
	if err != nil {
		return err
	}
	resp := sitedoctor.Run(context.Background(), site.Path)
	renderSiteDoctor(w, useColor, site, resp)
	return nil
}

// findSite looks a site up by its registered name, returning a helpful error
// that lists the known names when there is no match.
func findSite(sites []config.Site, name string) (config.Site, error) {
	known := make([]string, 0, len(sites))
	for _, s := range sites {
		if s.Name == name {
			return s, nil
		}
		known = append(known, s.Name)
	}
	sort.Strings(known)
	return config.Site{}, fmt.Errorf("no linked site named %q; known sites: %s", name, strings.Join(known, ", "))
}

func renderSiteDoctor(w io.Writer, useColor bool, site config.Site, resp sitedoctor.Response) {
	header := site.Name
	if len(site.Domains) > 0 {
		header = fmt.Sprintf("%s (%s)", site.Name, site.Domains[0])
	}
	fmt.Fprintf(w, "Site Doctor: %s\n", header)
	fmt.Fprintln(w, "══════════════════════════════════════════════")
	fmt.Fprintf(w, "  %-12s %s\n\n", "path", site.Path)

	if len(resp.Checks) == 0 {
		fmt.Fprintln(w, "  No app-level checks apply to this site (not a Laravel project, or nothing to flag).")
		return
	}

	for _, c := range resp.Checks {
		switch c.Status {
		case sitedoctor.StatusOK:
			fmt.Fprintf(w, "  %s %s\n", feedback.GreenIf(useColor, feedback.GlyphOK), c.Name)
		case sitedoctor.StatusWarn:
			fmt.Fprintf(w, "  %s %s  %s\n", feedback.AmberIf(useColor, feedback.GlyphWarn), c.Name, c.Detail)
		case sitedoctor.StatusFail:
			fmt.Fprintf(w, "  %s %s  %s\n", feedback.RedIf(useColor, feedback.GlyphFail), c.Name, c.Detail)
			if c.Fix != "" {
				fmt.Fprintf(w, "    hint: fix command \"%s\"\n", c.Fix)
			}
		default:
			fmt.Fprintf(w, "  - %s  %s\n", c.Name, c.Detail)
		}
	}

	fmt.Fprintln(w, "\n══════════════════════════════════════════════")
	switch {
	case resp.Failures > 0 && resp.Warnings > 0:
		fmt.Fprintln(w, feedback.RedIf(useColor, fmt.Sprintf("%d failure(s), %d warning(s) found.", resp.Failures, resp.Warnings)))
	case resp.Failures > 0:
		fmt.Fprintln(w, feedback.RedIf(useColor, fmt.Sprintf("%d failure(s) found.", resp.Failures)))
	case resp.Warnings > 0:
		fmt.Fprintf(w, "%s  No failures.\n", feedback.AmberIf(useColor, fmt.Sprintf("%d warning(s) found.", resp.Warnings)))
	default:
		fmt.Fprintln(w, feedback.GreenIf(useColor, "All app checks passed."))
	}
}
