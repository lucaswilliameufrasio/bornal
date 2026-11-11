package kit

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Look lists wow-kit files and directories one level under each path.
// A path that itself ends in .wowkit is included. Symlinks are skipped.
func Look(paths []string) []Hit {
	seen := map[string]bool{}
	var hits []Hit
	add := func(path string) {
		clean := filepath.Clean(path)
		if seen[clean] || !strings.HasSuffix(clean, ".wowkit") {
			return
		}
		info, err := os.Lstat(clean)
		if err != nil || info.Mode()&os.ModeSymlink != 0 {
			return
		}
		seen[clean] = true
		hit := Hit{Path: clean, Size: info.Size()}
		if info.IsDir() {
			manifest, err := peekManifest(clean)
			if err != nil {
				return
			}
			hit.Label = manifest.Label
			hit.Created = manifest.CreatedAt
			hit.Size = 0
			for _, file := range manifest.Files {
				hit.Size += file.Size
			}
		} else {
			label, date, encrypted, err := Header(clean)
			if err != nil || !encrypted {
				return
			}
			hit.Label = label
			hit.Created = date
			hit.Encrypted = true
		}
		hits = append(hits, hit)
	}
	for _, path := range paths {
		if strings.TrimSpace(path) == "" {
			continue
		}
		add(path)
		entries, err := os.ReadDir(path)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".wowkit") {
				add(filepath.Join(path, entry.Name()))
			}
		}
	}
	sort.Slice(hits, func(i, j int) bool { return hits[i].Path < hits[j].Path })
	return hits
}

// Places are the shallow directories a search is allowed to open.
func Places(home string, volumes []string, last string) []string {
	var out []string
	if last != "" {
		out = append(out, last, filepath.Dir(last))
	}
	if home != "" {
		out = append(out,
			filepath.Join(home, "WoW Kits"),
			filepath.Join(home, "Backups", "wow-kits"),
			filepath.Join(home, ".local", "share", "bornal", "kits"),
		)
	}
	for _, volume := range volumes {
		if strings.TrimSpace(volume) == "" {
			continue
		}
		out = append(out,
			volume,
			filepath.Join(volume, "WoW Kits"),
			filepath.Join(volume, "Backups", "wow-kits"),
		)
	}
	return out
}
