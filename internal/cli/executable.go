package cli

import (
	"os"
	"path/filepath"
	"strings"
)

func CurrentExecutablePath() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return ResolveExecutablePath(executable, home)
}

// ResolveExecutablePath returns canonical absolute executable path with home
// collapsed for compact, portable display.
func ResolveExecutablePath(executable, home string) (string, error) {
	absolute, err := filepath.Abs(executable)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return "", err
	}
	absoluteHome, err := filepath.Abs(home)
	if err != nil {
		return "", err
	}
	resolvedHome, err := filepath.EvalSymlinks(absoluteHome)
	if err != nil {
		return "", err
	}
	return CollapseHomePath(resolved, resolvedHome), nil
}

func CollapseHomePath(path, home string) string {
	path = filepath.Clean(path)
	home = filepath.Clean(home)
	if path == home {
		return "~"
	}
	prefix := home + string(filepath.Separator)
	if strings.HasPrefix(path, prefix) {
		return "~" + string(filepath.Separator) + strings.TrimPrefix(path, prefix)
	}
	return path
}
