package process

import (
	"os"
	"path/filepath"
	"strings"
)

func AnyRunning(executables []string) bool {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return false
	}
	wanted := make(map[string]bool, len(executables))
	for _, v := range executables {
		wanted[filepath.Base(v)] = true
	}
	for _, e := range entries {
		if !e.IsDir() || e.Name() == "self" {
			continue
		}
		if _, err := parsePID(e.Name()); err != nil {
			continue
		}
		if b, err := os.ReadFile(filepath.Join("/proc", e.Name(), "comm")); err == nil && wanted[strings.TrimSpace(string(b))] {
			return true
		}
		if target, err := os.Readlink(filepath.Join("/proc", e.Name(), "exe")); err == nil && wanted[filepath.Base(target)] {
			return true
		}
	}
	return false
}
func parsePID(s string) (int, error) {
	n := 0
	if s == "" {
		return 0, os.ErrInvalid
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, os.ErrInvalid
		}
		n = n*10 + int(r-'0')
	}
	return n, nil
}
