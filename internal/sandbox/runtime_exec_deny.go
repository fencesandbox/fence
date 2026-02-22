package sandbox

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Use-Tusk/fence/internal/config"
)

var commonExecutableDirs = []string{
	"/usr/bin",
	"/bin",
	"/usr/local/bin",
	"/opt/homebrew/bin",
	"/opt/local/bin",
}

// GetRuntimeDeniedExecutablePaths returns absolute executable paths that should
// be blocked at exec-time for this config.
//
// Runtime exec enforcement is intentionally conservative:
// - Only deny entries that are a single executable token are included.
// - Prefix rules with arguments (e.g. "git push", "dd if=") remain preflight-only.
func GetRuntimeDeniedExecutablePaths(cfg *config.Config) []string {
	if cfg == nil {
		return nil
	}

	var denyRules []string
	denyRules = append(denyRules, cfg.Command.Deny...)
	if cfg.Command.UseDefaultDeniedCommands() {
		denyRules = append(denyRules, config.DefaultDeniedCommands...)
	}

	var paths []string
	seen := make(map[string]bool)

	for _, rule := range denyRules {
		token, ok := runtimeExecutableToken(rule)
		if !ok {
			continue
		}

		for _, resolved := range resolveExecutablePaths(token) {
			if seen[resolved] {
				continue
			}
			seen[resolved] = true
			paths = append(paths, resolved)
		}
	}

	slices.Sort(paths)
	return paths
}

func runtimeExecutableToken(rule string) (string, bool) {
	rule = strings.TrimSpace(rule)
	if rule == "" {
		return "", false
	}

	tokens := tokenizeCommand(rule)
	if len(tokens) != 1 {
		return "", false
	}

	token := strings.TrimSpace(tokens[0])
	if token == "" {
		return "", false
	}

	// Runtime exec enforcement is path/name-based; skip entries that clearly
	// encode shell-level matching syntax.
	if strings.ContainsAny(token, "*?[]|&;()<>$`=") {
		return "", false
	}

	return token, true
}

func resolveExecutablePaths(token string) []string {
	var paths []string
	seen := make(map[string]bool)
	add := func(p string) {
		if p == "" || seen[p] {
			return
		}
		seen[p] = true
		paths = append(paths, p)
	}

	addCanonicalPath := func(p string) {
		if p == "" {
			return
		}
		add(p)
		if resolved, err := filepath.EvalSymlinks(p); err == nil {
			add(resolved)
		}
	}

	if strings.ContainsRune(token, filepath.Separator) {
		abs := token
		if !filepath.IsAbs(abs) {
			if cwd, err := os.Getwd(); err == nil {
				abs = filepath.Join(cwd, abs)
			}
		}
		if executablePathExists(abs) {
			addCanonicalPath(abs)
		}
		return paths
	}

	if resolved, err := exec.LookPath(token); err == nil {
		addCanonicalPath(resolved)
	}

	for _, dir := range commonExecutableDirs {
		candidate := filepath.Join(dir, token)
		if executablePathExists(candidate) {
			addCanonicalPath(candidate)
		}
	}

	return paths
}

func executablePathExists(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
