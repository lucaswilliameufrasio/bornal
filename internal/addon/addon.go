// Package addon reads addon TOC files. It does not install addons or run Lua.
package addon

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Addon is one folder under Interface/AddOns.
type Addon struct {
	Folder       string   `json:"folder"`
	Title        string   `json:"title,omitempty"`
	Version      string   `json:"version,omitempty"`
	Interface    string   `json:"interface,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"`
	Missing      []string `json:"missing,omitempty"`
	Disabled     bool     `json:"disabled,omitempty"`
	InterfaceOK  bool     `json:"interface_ok"`
}

// List reads TOC metadata for one flavor directory. A missing AddOns folder is an empty list.
func List(flavor string) ([]Addon, error) {
	root := filepath.Join(flavor, "Interface", "AddOns")
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return []Addon{}, nil
		}
		return nil, err
	}
	disabled := disabledNames(flavor)
	found := make([]Addon, 0)
	names := map[string]bool{}
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		info, statErr := os.Lstat(filepath.Join(root, entry.Name()))
		if statErr != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			continue
		}
		names[entry.Name()] = true
		one, ok := readTOC(root, entry.Name())
		if !ok {
			continue
		}
		one.Disabled = disabled[one.Folder]
		found = append(found, one)
	}
	for i := range found {
		for _, dep := range found[i].Dependencies {
			if !names[dep] {
				found[i].Missing = append(found[i].Missing, dep)
			}
		}
	}
	sort.Slice(found, func(i, j int) bool {
		return found[i].Folder < found[j].Folder
	})
	return found, nil
}

func readTOC(root, folder string) (Addon, bool) {
	path := filepath.Join(root, folder, folder+".toc")
	file, err := os.Open(path)
	if err != nil {
		matches, _ := filepath.Glob(filepath.Join(root, folder, "*.toc"))
		if len(matches) == 0 {
			return Addon{}, false
		}
		file, err = os.Open(matches[0])
		if err != nil {
			return Addon{}, false
		}
	}
	defer file.Close()
	addon := Addon{Folder: folder, InterfaceOK: true}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "##") {
			continue
		}
		body := strings.TrimSpace(strings.TrimPrefix(line, "##"))
		key, value, ok := strings.Cut(body, ":")
		if !ok {
			continue
		}
		key = strings.ToLower(strings.TrimSpace(key))
		value = strings.TrimSpace(value)
		switch key {
		case "title":
			addon.Title = value
		case "version":
			addon.Version = value
		case "interface":
			addon.Interface = value
		case "dependencies", "requireddeps":
			addon.Dependencies = splitList(value)
		}
	}
	if err := scanner.Err(); err != nil {
		return Addon{}, false
	}
	if addon.Title == "" {
		addon.Title = folder
	}
	if addon.Interface != "" && filepath.Base(filepath.Dir(filepath.Dir(root))) == "_forever_" && addon.Interface != "16001" {
		addon.InterfaceOK = false
	}
	return addon, true
}

func splitList(value string) []string {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ' '
	})
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func disabledNames(flavor string) map[string]bool {
	found := map[string]bool{}
	accountRoot := filepath.Join(flavor, "WTF", "Account")
	accounts, err := os.ReadDir(accountRoot)
	if err != nil {
		return found
	}
	for _, account := range accounts {
		if !account.IsDir() {
			continue
		}
		accountPath := filepath.Join(accountRoot, account.Name())
		readDisabled(filepath.Join(accountPath, "AddOns.txt"), found)
		realms, realmErr := os.ReadDir(accountPath)
		if realmErr != nil {
			continue
		}
		for _, realm := range realms {
			if !realm.IsDir() || realm.Name() == "SavedVariables" {
				continue
			}
			chars, charErr := os.ReadDir(filepath.Join(accountPath, realm.Name()))
			if charErr != nil {
				continue
			}
			for _, char := range chars {
				if !char.IsDir() {
					continue
				}
				readDisabled(filepath.Join(accountPath, realm.Name(), char.Name(), "AddOns.txt"), found)
			}
		}
	}
	return found
}

func readDisabled(path string, found map[string]bool) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		name := line
		state := ""
		if cut, rest, ok := strings.Cut(line, ":"); ok {
			name = strings.TrimSpace(cut)
			state = strings.ToLower(strings.TrimSpace(rest))
		}
		if name != "" && strings.Contains(state, "disabled") {
			found[name] = true
		}
	}
}
