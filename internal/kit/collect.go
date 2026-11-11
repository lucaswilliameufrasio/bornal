package kit

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/lucaswilliameufrasio/bornal/internal/addon"
)

var secretLine = regexp.MustCompile(`(?i)(?:^|[^A-Za-z0-9_])(password|passwd|api[_-]?key|secret|token|authkey)\s*=`)

type stored struct {
	Rel  string
	Data []byte
}

func collect(installPath, account string) ([]stored, []Character, []string, []Warning, string, error) {
	accountPath := filepath.Join(installPath, "WTF", "Account", account)
	info, err := os.Lstat(accountPath)
	if err != nil {
		return nil, nil, nil, nil, "", err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, nil, nil, nil, "", ErrAccount
	}
	var files []stored
	var total int
	var warnings []Warning
	var omitted []string
	noteOmitted(accountPath, &omitted)

	for _, name := range []string{"macros-cache.txt", "bindings-cache.wtf"} {
		if err := addNamed(&files, &total, &warnings, accountPath, "account/"+name, name); err != nil {
			return nil, nil, nil, nil, "", err
		}
	}
	entries, err := os.ReadDir(accountPath)
	if err != nil {
		return nil, nil, nil, nil, "", err
	}
	for _, entry := range entries {
		if entry.IsDir() || !editLike(entry.Name()) || omitName(entry.Name()) {
			continue
		}
		if err := addNamed(&files, &total, &warnings, accountPath, "account/"+entry.Name(), entry.Name()); err != nil {
			return nil, nil, nil, nil, "", err
		}
	}
	if err := walkSaved(filepath.Join(accountPath, "SavedVariables"), "account/SavedVariables", &files, &total, &warnings, &omitted); err != nil {
		return nil, nil, nil, nil, "", err
	}

	var characters []Character
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == "SavedVariables" || omitName(entry.Name()) {
			continue
		}
		if !utf8.ValidString(entry.Name()) || !segment(entry.Name()) {
			omitted = append(omitted, "realm")
			continue
		}
		realmPath := filepath.Join(accountPath, entry.Name())
		realmInfo, statErr := os.Lstat(realmPath)
		if statErr != nil || !realmInfo.IsDir() || realmInfo.Mode()&os.ModeSymlink != 0 {
			continue
		}
		chars, readErr := os.ReadDir(realmPath)
		if readErr != nil {
			return nil, nil, nil, nil, "", readErr
		}
		for _, char := range chars {
			if !char.IsDir() || omitName(char.Name()) || !segment(char.Name()) || !utf8.ValidString(char.Name()) {
				continue
			}
			charPath := filepath.Join(realmPath, char.Name())
			charInfo, charErr := os.Lstat(charPath)
			if charErr != nil || !charInfo.IsDir() || charInfo.Mode()&os.ModeSymlink != 0 {
				continue
			}
			relBase := path.Join("characters", entry.Name(), char.Name())
			if err := addNamed(&files, &total, &warnings, charPath, relBase+"/AddOns.txt", "AddOns.txt"); err != nil {
				return nil, nil, nil, nil, "", err
			}
			charEntries, listErr := os.ReadDir(charPath)
			if listErr != nil {
				return nil, nil, nil, nil, "", listErr
			}
			for _, file := range charEntries {
				if file.IsDir() || !editLike(file.Name()) || omitName(file.Name()) {
					continue
				}
				if err := addNamed(&files, &total, &warnings, charPath, relBase+"/"+file.Name(), file.Name()); err != nil {
					return nil, nil, nil, nil, "", err
				}
			}
			if err := walkSaved(filepath.Join(charPath, "SavedVariables"), relBase+"/SavedVariables", &files, &total, &warnings, &omitted); err != nil {
				return nil, nil, nil, nil, "", err
			}
			characters = append(characters, Character{Name: char.Name(), Realm: entry.Name()})
		}
	}

	configBody, err := portableBody(filepath.Join(installPath, "WTF", "Config.wtf"))
	if err != nil {
		return nil, nil, nil, nil, "", err
	}
	if len(configBody) > 0 {
		if err := addBytes(&files, &total, &warnings, "config/portable.wtf", configBody); err != nil {
			return nil, nil, nil, nil, "", err
		}
	}

	listed, err := addon.List(installPath)
	if err != nil {
		return nil, nil, nil, nil, "", err
	}
	index, err := json.MarshalIndent(listed, "", "  ")
	if err != nil {
		return nil, nil, nil, nil, "", err
	}
	index = append(index, '\n')
	if err := addBytes(&files, &total, &warnings, "addons/index.json", index); err != nil {
		return nil, nil, nil, nil, "", err
	}
	interfaceID := commonInterface(listed)

	sort.Slice(files, func(i, j int) bool { return files[i].Rel < files[j].Rel })
	sort.Slice(characters, func(i, j int) bool {
		if characters[i].Realm == characters[j].Realm {
			return characters[i].Name < characters[j].Name
		}
		return characters[i].Realm < characters[j].Realm
	})
	sort.Slice(warnings, func(i, j int) bool {
		if warnings[i].File == warnings[j].File {
			return warnings[i].Line < warnings[j].Line
		}
		return warnings[i].File < warnings[j].File
	})
	omitted = unique(omitted)
	return files, characters, omitted, warnings, interfaceID, nil
}

func addNamed(files *[]stored, total *int, warnings *[]Warning, dir, rel, name string) error {
	path := filepath.Join(dir, name)
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return nil
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return addBytes(files, total, warnings, rel, body)
}

func walkSaved(abs, rel string, files *[]stored, total *int, warnings *[]Warning, omitted *[]string) error {
	info, err := os.Lstat(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil
	}
	return walkFiles(abs, rel, files, total, warnings, omitted)
}

func walkFiles(abs, rel string, files *[]stored, total *int, warnings *[]Warning, omitted *[]string) error {
	entries, err := os.ReadDir(abs)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		name := entry.Name()
		nextAbs := filepath.Join(abs, name)
		nextRel := path.Join(rel, name)
		info, statErr := os.Lstat(nextAbs)
		if statErr != nil {
			return statErr
		}
		if info.Mode()&os.ModeSymlink != 0 {
			*omitted = append(*omitted, name)
			continue
		}
		if omitName(name) {
			*omitted = append(*omitted, name)
			continue
		}
		if info.IsDir() {
			if err := walkFiles(nextAbs, nextRel, files, total, warnings, omitted); err != nil {
				return err
			}
			continue
		}
		if !info.Mode().IsRegular() {
			continue
		}
		body, err := os.ReadFile(nextAbs)
		if err != nil {
			return err
		}
		if err := addBytes(files, total, warnings, nextRel, body); err != nil {
			return err
		}
	}
	return nil
}

func addBytes(files *[]stored, total *int, warnings *[]Warning, rel string, body []byte) error {
	clean, err := safeSlash(rel)
	if err != nil {
		return err
	}
	if len(*files) >= maxCount || len(body) > maxFile || *total+len(body) > maxTotal {
		return ErrTooBig
	}
	*total += len(body)
	*files = append(*files, stored{Rel: clean, Data: body})
	*warnings = append(*warnings, secretWarnings(clean, body)...)
	return nil
}

func secretWarnings(rel string, body []byte) []Warning {
	if !strings.HasSuffix(strings.ToLower(rel), ".lua") {
		return nil
	}
	var out []Warning
	lines := bytes.Split(body, []byte("\n"))
	for i, line := range lines {
		if secretLine.Match(line) {
			out = append(out, Warning{File: rel, Line: i + 1})
		}
	}
	return out
}

func omitName(name string) bool {
	lower := strings.ToLower(name)
	switch lower {
	case "cache", "logs", "data", "agent", "blizzard_console.lua", "macros-cache.old", "chat":
		return true
	}
	if strings.HasSuffix(lower, ".bak") || strings.HasSuffix(lower, ".old") || strings.HasSuffix(lower, ".bornal-bak") || strings.HasSuffix(lower, ".bornal-tmp") {
		return true
	}
	if strings.HasPrefix(lower, "chat") {
		return true
	}
	return false
}

func editLike(name string) bool {
	lower := strings.ToLower(name)
	return strings.Contains(lower, "edit") || strings.Contains(lower, "layout")
}

func segment(name string) bool {
	return name != "" && name != "." && name != ".." && !strings.ContainsAny(name, `/\`)
}

func noteOmitted(accountPath string, omitted *[]string) {
	for _, name := range []string{"Cache", "Logs", "Data", "Agent"} {
		if _, err := os.Lstat(filepath.Join(accountPath, name)); err == nil {
			*omitted = append(*omitted, name)
		}
	}
}

func portableBody(configPath string) ([]byte, error) {
	file, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()
	var buf bytes.Buffer
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		key, value, ok := parseSet(scanner.Text())
		if !ok || !allowKey(key) {
			continue
		}
		buf.WriteString("SET ")
		buf.WriteString(key)
		buf.WriteString(" \"")
		buf.WriteString(value)
		buf.WriteString("\"\n")
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func parseSet(line string) (string, string, bool) {
	fields := strings.Fields(line)
	if len(fields) < 3 || !strings.EqualFold(fields[0], "SET") {
		return "", "", false
	}
	value := strings.Trim(strings.Join(fields[2:], " "), `"`)
	return fields[1], value, true
}

func allowKey(key string) bool {
	switch strings.ToLower(key) {
	case "textlocale", "audiolocale", "locale":
		return true
	default:
		return false
	}
}

func commonInterface(list []addon.Addon) string {
	counts := map[string]int{}
	best := ""
	bestN := 0
	for _, one := range list {
		if one.Interface == "" {
			continue
		}
		counts[one.Interface]++
		if counts[one.Interface] > bestN {
			bestN = counts[one.Interface]
			best = one.Interface
		}
	}
	return best
}

func unique(items []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		out = append(out, item)
	}
	sort.Strings(out)
	return out
}
