package dns

import (
	"strings"
	"testing"
)

// assertNoWildcardArgs walks the sudoers drop-in line by line and fails the
// test if any command argument (token after the verb) contains a sudo
// wildcard character. sudo >= 1.9.16 hard-rejects wildcards in command
// arguments and falls back to the password-prompt path on every call,
// which is what bug #269 reported. Catching this in CI prevents a
// regression that would silently break installs on strict-sudo distros
// (Ubuntu 26.04, Fedora 41+, Arch / CachyOS, openSUSE Tumbleweed, etc).
func assertNoWildcardArgs(t *testing.T, content string) {
	t.Helper()
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Sudoers rule: <user> ALL=(root) NOPASSWD: /path/to/cmd args...
		idx := strings.Index(line, "NOPASSWD:")
		if idx < 0 {
			continue
		}
		cmd := strings.TrimSpace(line[idx+len("NOPASSWD:"):])
		// Multiple commands can appear on one line separated by ", ".
		for _, c := range strings.Split(cmd, ", ") {
			tokens := strings.Fields(c)
			if len(tokens) <= 1 {
				continue
			}
			for _, arg := range tokens[1:] {
				if strings.ContainsAny(arg, "*?") {
					t.Errorf("wildcard in sudoers command argument: %q (full line: %q)", arg, line)
				}
			}
		}
	}
}

func TestRenderLinuxSudoers_NoWildcardArgs(t *testing.T) {
	content := renderLinuxSudoers("alice")
	assertNoWildcardArgs(t, content)
}

func TestRenderLinuxSudoers_IncludesUserOnEveryRule(t *testing.T) {
	content := renderLinuxSudoers("alice")
	rules := 0
	for _, line := range strings.Split(content, "\n") {
		if strings.Contains(line, "NOPASSWD:") {
			rules++
			if !strings.HasPrefix(line, "alice ") {
				t.Errorf("rule does not start with the user: %q", line)
			}
		}
	}
	if rules == 0 {
		t.Fatal("expected at least one sudoers rule, got none")
	}
}
