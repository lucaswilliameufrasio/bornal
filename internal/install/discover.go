package install

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

const (
	maxSearchDepth   = 4
	maxSearchEntries = 2000
)

// heavy is a directory name the search does not enter.
var heavy = map[string]bool{
	"Data": true, "Cache": true, "Interface": true, "node_modules": true,
	".git": true, "target": true, "dist": true, "WTF": true,
}

// Discover returns installs under the given roots. Each root is either an
// install or a place to look a few levels down. The walk is capped.
func Discover(roots []string) []Install {
	seen := map[string]bool{}
	found := make([]Install, 0)
	visited := 0
	for _, root := range roots {
		if root == "" {
			continue
		}
		walk(filepath.Clean(root), 0, &visited, seen, &found)
	}
	sort.Slice(found, func(i, j int) bool {
		return found[i].Path < found[j].Path
	})
	return found
}

// DefaultRoots are the places the first-run question looks. The walk under
// each root is capped by Discover.
func DefaultRoots(getenv func(string) string) []string {
	return RootsFor(runtime.GOOS, getenv)
}

// RootsFor lists search roots for one operating system. It does not touch the disk.
func RootsFor(goos string, getenv func(string) string) []string {
	roots := make([]string, 0)
	if getenv == nil {
		return roots
	}
	if path := strings.TrimSpace(getenv("BORNAL_WOW_PATH")); path != "" {
		roots = append(roots, path)
	}
	home := getenv("HOME")
	if home == "" {
		home = getenv("USERPROFILE")
	}
	if home != "" {
		roots = append(roots, filepath.Join(home, "Games"))
	}
	switch goos {
	case "darwin":
		roots = append(roots, "/Applications", "/Volumes")
	case "windows":
		if path := strings.TrimSpace(getenv("ProgramFiles")); path != "" {
			roots = append(roots, path)
		}
		if path := strings.TrimSpace(getenv("ProgramFiles(x86)")); path != "" {
			roots = append(roots, path)
		}
	default:
		roots = append(roots, "/mnt", "/media")
		user := getenv("USER")
		if user == "" {
			user = getenv("USERNAME")
		}
		if user != "" {
			roots = append(roots, filepath.Join("/run/media", user))
		}
	}
	return roots
}

func walk(dir string, depth int, visited *int, seen map[string]bool, found *[]Install) {
	if *visited >= maxSearchEntries || depth > maxSearchDepth {
		return
	}
	*visited++
	one, ok, err := readFlavor(dir)
	if err == nil && ok {
		if !seen[one.Path] {
			seen[one.Path] = true
			*found = append(*found, one)
		}
		return
	}
	if depth == maxSearchDepth {
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if *visited >= maxSearchEntries {
			return
		}
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if heavy[name] || (len(name) > 0 && name[0] == '.') {
			continue
		}
		walk(filepath.Join(dir, name), depth+1, visited, seen, found)
	}
}
