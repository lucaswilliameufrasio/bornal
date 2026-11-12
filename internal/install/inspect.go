// Package install reads a World of Warcraft directory without writing it.
package install

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Install is one flavor directory, such as _forever_.
type Install struct {
	Path     string
	Flavor   string
	Locale   string
	AgentUID string
}

// Inspect lists flavor directories under root. root itself counts when it
// holds WTF/Config.wtf and no child does. The walk stops at that child.
func Inspect(root string) ([]Install, error) {
	cleaned := filepath.Clean(root)
	info, err := os.Stat(cleaned)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s não é um diretório", cleaned)
	}
	children, err := flavorsUnder(cleaned)
	if err != nil {
		return nil, err
	}
	if len(children) > 0 {
		sort.Slice(children, func(i, j int) bool { return children[i].Path < children[j].Path })
		return children, nil
	}
	one, ok, err := readFlavor(cleaned)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	return []Install{one}, nil
}

func flavorsUnder(root string) ([]Install, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	found := make([]Install, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		one, ok, err := readFlavor(filepath.Join(root, entry.Name()))
		if err != nil {
			return nil, err
		}
		if ok {
			found = append(found, one)
		}
	}
	return found, nil
}

func readFlavor(dir string) (Install, bool, error) {
	configPath := filepath.Join(dir, "WTF", "Config.wtf")
	file, err := os.Open(configPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Install{}, false, nil
		}
		return Install{}, false, err
	}
	defer file.Close()

	install := Install{Path: dir, Flavor: filepath.Base(dir)}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		key, value, ok := setLine(scanner.Text())
		if !ok {
			continue
		}
		switch strings.ToLower(key) {
		case "textlocale":
			install.Locale = value
		case "agentuid":
			install.AgentUID = value
		}
	}
	if err := scanner.Err(); err != nil {
		return Install{}, false, err
	}
	return install, true, nil
}

func setLine(line string) (string, string, bool) {
	fields := strings.Fields(line)
	if len(fields) < 3 || !strings.EqualFold(fields[0], "SET") {
		return "", "", false
	}
	value := strings.Join(fields[2:], " ")
	value = strings.Trim(value, `"`)
	return fields[1], value, true
}
