// Package running reports whether Wow.exe for one install is open.
package running

import (
	"os"
	"path/filepath"
	"strings"
)

// Process is one process the scanner could read.
type Process struct {
	Exe     string
	Cmdline string
}

// List is the process scan. Tests replace it.
var List = scan

// Open reports whether a process refers to this install. A check that cannot
// run returns an error so the caller refuses to write.
func Open(install string) (bool, error) {
	if List == nil {
		return false, os.ErrInvalid
	}
	procs, err := List()
	if err != nil {
		return false, err
	}
	return Matches(install, procs), nil
}

// Matches reports whether any process text refers to this install.
func Matches(install string, procs []Process) bool {
	for _, proc := range procs {
		if refers(install, proc.Exe) || refers(install, proc.Cmdline) {
			return true
		}
	}
	return false
}

func refers(install, text string) bool {
	if strings.TrimSpace(text) == "" {
		return false
	}
	clean := filepath.Clean(install)
	folded := strings.ReplaceAll(text, "\\", "/")
	if strings.Contains(folded, clean) {
		return true
	}
	tail := wowTail(clean)
	lower := strings.ToLower(folded)
	if tail == "" || !strings.Contains(lower, "wow.exe") {
		return false
	}
	return strings.Contains(lower, strings.ToLower(tail))
}

func wowTail(install string) string {
	flavor := filepath.Base(install)
	parent := filepath.Base(filepath.Dir(install))
	if flavor == "." || flavor == string(filepath.Separator) {
		return ""
	}
	if strings.EqualFold(parent, "World of Warcraft") {
		return parent + "/" + flavor
	}
	return flavor
}

func scan() ([]Process, error) {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, os.ErrInvalid
		}
		return nil, err
	}
	found := make([]Process, 0)
	for _, entry := range entries {
		name := entry.Name()
		if name == "" || name[0] < '1' || name[0] > '9' {
			continue
		}
		dir := filepath.Join("/proc", name)
		exe, _ := os.Readlink(filepath.Join(dir, "exe"))
		raw, readErr := os.ReadFile(filepath.Join(dir, "cmdline"))
		if readErr != nil && exe == "" {
			continue
		}
		text := strings.ReplaceAll(string(raw), "\x00", " ")
		if !strings.Contains(strings.ToLower(exe+" "+text), "wow.exe") {
			continue
		}
		found = append(found, Process{Exe: exe, Cmdline: text})
	}
	return found, nil
}
