package kit

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// RestoreOptions copies a verified kit into an existing account folder.
type RestoreOptions struct {
	KitPath     string
	InstallPath string
	Locale      string
	Build       string
	Passphrase  []byte
	Replace     bool
	AcceptRealm bool
	Work        int
}

// Restore checks every hash before it writes. A failed check leaves the destination unchanged.
func Restore(opts RestoreOptions) (RestoreReport, error) {
	dir, cleanup, report, err := openKit(opts.KitPath, opts.Passphrase)
	defer cleanup()
	if err != nil {
		return RestoreReport{}, err
	}
	return restoreDir(dir, report.Manifest, opts)
}

func restoreDir(dir string, manifest Manifest, opts RestoreOptions) (RestoreReport, error) {
	if filepath.Base(opts.InstallPath) != manifest.Flavor && manifest.Flavor != "" {
		return RestoreReport{}, ErrFlavor
	}
	if manifest.Locale != "" && opts.Locale != "" && !sameLocale(manifest.Locale, opts.Locale) {
		return RestoreReport{}, ErrLocale
	}
	if manifest.Build != "" && opts.Build != "" && manifest.Build != opts.Build {
		return RestoreReport{}, ErrBuild
	}
	accountDir := filepath.Join(opts.InstallPath, "WTF", "Account", manifest.AccountID)
	info, err := os.Lstat(accountDir)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return RestoreReport{}, ErrAccount
	}
	mismatch, missing := realmIssues(accountDir, manifest.Characters)
	if len(mismatch) > 0 && !opts.AcceptRealm {
		return RestoreReport{SkippedCharacters: mismatch}, ErrRealm
	}
	when := created(manifest)
	var planned []plannedFile
	var skipped []string
	skipChar := map[string]bool{}
	for _, name := range missing {
		skipChar[name] = true
		skipped = append(skipped, name)
	}
	if !opts.AcceptRealm {
		for _, name := range mismatch {
			skipChar[name] = true
		}
	}
	for _, file := range manifest.Files {
		if strings.HasPrefix(file.Path, "addons/") {
			continue
		}
		if char, ok := characterOf(file.Path); ok && skipChar[char] {
			continue
		}
		target, err := destPath(opts.InstallPath, manifest.AccountID, file.Path)
		if err != nil {
			return RestoreReport{}, err
		}
		if target == "" {
			continue
		}
		if newer, err := isNewer(target, when); err != nil {
			return RestoreReport{}, err
		} else if newer && !opts.Replace {
			return RestoreReport{}, ErrNewer
		}
		source, err := joinUnder(dir, file.Path)
		if err != nil {
			return RestoreReport{}, err
		}
		body, err := os.ReadFile(source)
		if err != nil {
			return RestoreReport{}, err
		}
		if hashOf(body) != file.SHA256 {
			return RestoreReport{}, ErrHash
		}
		planned = append(planned, plannedFile{Path: target, Data: body, Rel: file.Path})
	}
	var portable []byte
	for i, file := range planned {
		if file.Rel == "config/portable.wtf" {
			portable = file.Data
			planned = append(planned[:i], planned[i+1:]...)
			break
		}
	}
	for _, file := range planned {
		if filepath.Base(file.Path) == "macros-cache.txt" {
			if err := copyBackup(file.Path); err != nil {
				return RestoreReport{}, err
			}
		}
		if err := writeAtomic(file.Path, file.Data, 0o644); err != nil {
			return RestoreReport{}, err
		}
	}
	if len(portable) > 0 {
		if err := mergeConfig(filepath.Join(opts.InstallPath, "WTF", "Config.wtf"), portable); err != nil {
			return RestoreReport{}, err
		}
	}
	sort.Strings(skipped)
	return RestoreReport{Written: len(planned) + boolCount(len(portable) > 0), SkippedCharacters: skipped, Warnings: manifest.Warnings}, nil
}

type plannedFile struct {
	Path string
	Rel  string
	Data []byte
}

func boolCount(ok bool) int {
	if ok {
		return 1
	}
	return 0
}

func sameLocale(a, b string) bool {
	return strings.EqualFold(strings.ReplaceAll(a, "-", ""), strings.ReplaceAll(b, "-", ""))
}

func characterOf(rel string) (string, bool) {
	parts := strings.Split(rel, "/")
	if len(parts) < 3 || parts[0] != "characters" {
		return "", false
	}
	return parts[2], true
}

func destPath(install, account, rel string) (string, error) {
	switch {
	case strings.HasPrefix(rel, "account/"):
		return joinUnder(filepath.Join(install, "WTF", "Account", account), strings.TrimPrefix(rel, "account/"))
	case strings.HasPrefix(rel, "characters/"):
		rest := strings.TrimPrefix(rel, "characters/")
		return joinUnder(filepath.Join(install, "WTF", "Account", account), rest)
	case rel == "config/portable.wtf":
		return filepath.Join(install, "WTF", "Config.wtf"), nil
	default:
		return "", ErrEscape
	}
}

func isNewer(path string, kitTime time.Time) (bool, error) {
	info, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if kitTime.IsZero() {
		return false, nil
	}
	return info.ModTime().After(kitTime), nil
}

func realmIssues(accountDir string, characters []Character) (mismatch, missing []string) {
	for _, character := range characters {
		realmPath := filepath.Join(accountDir, character.Realm)
		info, err := os.Lstat(realmPath)
		realmOK := err == nil && info.IsDir() && info.Mode()&os.ModeSymlink == 0
		elsewhere := false
		entries, readErr := os.ReadDir(accountDir)
		if readErr == nil {
			for _, entry := range entries {
				if !entry.IsDir() || entry.Name() == "SavedVariables" || entry.Name() == character.Realm {
					continue
				}
				charInfo, statErr := os.Lstat(filepath.Join(accountDir, entry.Name(), character.Name))
				if statErr == nil && charInfo.IsDir() {
					elsewhere = true
				}
			}
		}
		if !realmOK {
			missing = append(missing, character.Name)
		}
		if elsewhere {
			mismatch = append(mismatch, character.Name)
		}
	}
	return mismatch, missing
}

func copyBackup(path string) error {
	body, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return os.WriteFile(path+".bornal-bak", body, 0o644)
}

func mergeConfig(path string, portable []byte) error {
	original, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	newline := "\n"
	if bytes.Contains(original, []byte("\r\n")) {
		newline = "\r\n"
	}
	text := strings.ReplaceAll(string(original), "\r\n", "\n")
	updates := map[string]string{}
	var keys []string
	for _, line := range strings.Split(string(portable), "\n") {
		key, value, ok := parseSet(line)
		if !ok || !allowKey(key) {
			continue
		}
		lower := strings.ToLower(key)
		if _, seen := updates[lower]; !seen {
			keys = append(keys, lower)
		}
		updates[lower] = value
	}
	lines := strings.Split(text, "\n")
	seen := map[string]bool{}
	for i, line := range lines {
		key, _, ok := parseSet(line)
		if !ok {
			continue
		}
		lower := strings.ToLower(key)
		value, ok := updates[lower]
		if !ok {
			continue
		}
		lines[i] = "SET " + key + " \"" + value + "\""
		seen[lower] = true
	}
	for _, key := range keys {
		if seen[key] {
			continue
		}
		lines = append(lines, "SET "+key+" \""+updates[key]+"\"")
	}
	body := strings.Join(lines, newline)
	if !strings.HasSuffix(body, newline) {
		body += newline
	}
	return writeAtomic(path, []byte(body), 0o644)
}
