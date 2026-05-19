package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/geodro/lerd/internal/config"
	nodeDet "github.com/geodro/lerd/internal/node"
	"github.com/spf13/cobra"
)

// NewNodeCmd returns the node command.
func NewNodeCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "node [args...]",
		Short:              "Run node using the project's version via fnm",
		DisableFlagParsing: true,
		SilenceUsage:       true,
		RunE: func(_ *cobra.Command, args []string) error {
			return runWithFnm("node", args)
		},
	}
}

// NewNpmCmd returns the npm command.
func NewNpmCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "npm [args...]",
		Short:              "Run npm using the project's node version via fnm",
		DisableFlagParsing: true,
		SilenceUsage:       true,
		RunE: func(_ *cobra.Command, args []string) error {
			return runWithFnm("npm", args)
		},
	}
}

// NewNpxCmd returns the npx command.
func NewNpxCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "npx [args...]",
		Short:              "Run npx using the project's node version via fnm",
		DisableFlagParsing: true,
		SilenceUsage:       true,
		RunE: func(_ *cobra.Command, args []string) error {
			return runWithFnm("npx", args)
		},
	}
}

func runWithFnm(bin string, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	version, _ := nodeDet.DetectVersion(cwd)
	// Empty means the user has no .nvmrc / .node-version / global default; fall
	// through to the fnm `default` alias so we still surface a friendly error
	// instead of an unhelpful "Can't find version in dotfiles".
	if version == "" {
		version = "default"
	}

	fnm := filepath.Join(config.BinDir(), "fnm")
	if _, err := os.Stat(fnm); err != nil {
		return fmt.Errorf("fnm not found at %s — run 'lerd install' first", fnm)
	}

	if version != "default" {
		// Ensure the version is installed (suppress output — fnm prints even when already installed)
		_ = exec.Command(fnm, "install", version).Run()
	} else if exec.Command(fnm, "exec", "--using=default", "--", "true").Run() != nil {
		return fmt.Errorf("no Node.js version available via lerd — run: lerd node:install 22")
	}

	cmdArgs := []string{"exec", "--using=" + version, "--", bin}
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.Command(fnm, cmdArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Only npm/npx care about npm_config_prefix and produce installable bins,
	// so plain `lerd node script.js` skips the prefix override and the sync.
	manageGlobals := bin == "npm" || bin == "npx"
	prefix := config.NodeGlobalDir()
	if manageGlobals {
		if err := os.MkdirAll(filepath.Join(prefix, "bin"), 0o755); err == nil {
			cmd.Env = append(os.Environ(), "npm_config_prefix="+prefix)
		}
	}
	runErr := cmd.Run()
	if manageGlobals {
		// Sync wrappers regardless of exit status: an `npm uninstall -g` that
		// fails partway can still have removed a bin we want to mirror out of
		// ~/.local/bin/.
		if syncErr := syncNodeGlobalBins(filepath.Join(prefix, "bin"), config.BinDir(), fnm); syncErr != nil {
			fmt.Fprintf(os.Stderr, "lerd: warning: failed to sync npm global wrappers: %v\n", syncErr)
		}
	}
	if runErr != nil {
		if exit, ok := runErr.(*exec.ExitError); ok {
			os.Exit(exit.ExitCode())
		}
		return runErr
	}
	return nil
}

// shimMarker tags wrapper scripts lerd writes into ~/.local/bin/ for npm
// globals, so sync only ever removes its own files.
const shimMarker = "lerd-managed npm global shim"

// syncNodeGlobalBins mirrors every executable in sourceBin into targetBin as a
// tiny shell wrapper that exec's the real bin through `fnm exec`, so the
// shebang `#!/usr/bin/env node` resolves against the fnm-managed node.
//
// Foreign files (anything in targetBin without the marker) are left alone so
// we never clobber a tool the user installed by hand. Orphan wrappers (marker
// present, source gone) are removed.
func syncNodeGlobalBins(sourceBin, targetBin, fnmPath string) error {
	if err := os.MkdirAll(targetBin, 0o755); err != nil {
		return err
	}

	want := map[string]bool{}
	entries, err := os.ReadDir(sourceBin)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		want[name] = true
		wrapperPath := filepath.Join(targetBin, name)
		// Preserve foreign files: only overwrite if the existing file is
		// already a lerd-managed shim (or doesn't exist).
		if existing, ok := readShimHead(wrapperPath); ok && !strings.Contains(existing, shimMarker) {
			continue
		} else if !ok {
			// file exists but isn't a shell script (e.g. a Go binary); leave alone
			if _, statErr := os.Stat(wrapperPath); statErr == nil {
				continue
			}
		}
		realBin := filepath.Join(sourceBin, name)
		// --using=default so the wrapper works from any directory; without
		// it fnm errors out unless the cwd has a .nvmrc/.node-version.
		body := fmt.Sprintf("#!/bin/sh\n# %s\nexec %q exec --using=default -- %q \"$@\"\n", shimMarker, fnmPath, realBin)
		if err := os.WriteFile(wrapperPath, []byte(body), 0o755); err != nil {
			return err
		}
	}

	// Cleanup: remove our wrappers whose source bin no longer exists.
	targetEntries, err := os.ReadDir(targetBin)
	if err != nil {
		return err
	}
	for _, e := range targetEntries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if want[name] {
			continue
		}
		full := filepath.Join(targetBin, name)
		head, ok := readShimHead(full)
		if !ok {
			// not a shell script — could be the lerd binary itself, never touch
			continue
		}
		if !strings.Contains(head, shimMarker) {
			continue
		}
		_ = os.Remove(full)
	}
	return nil
}

// readShimHead returns the first ~256 bytes of path as a string, but only if
// the file starts with a shell shebang. The second return is false when the
// file isn't a shell script, so callers can skip native binaries without
// risking a false-positive marker match on string constants inside them.
func readShimHead(path string) (string, bool) {
	f, err := os.Open(path)
	if err != nil {
		return "", false
	}
	defer f.Close()
	buf := make([]byte, 256)
	n, _ := f.Read(buf)
	if n < 2 || buf[0] != '#' || buf[1] != '!' {
		return "", false
	}
	return string(buf[:n]), true
}
