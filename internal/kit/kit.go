// Package kit builds and checks a wow-kit v1.
// SavedVariables are copied as text. They are not executed.
package kit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	Format   = "wow-kit"
	Version  = 1
	maxFile  = 8 << 20
	maxTotal = 64 << 20
	maxCount = 500
)

var (
	ErrEscape  = errors.New("path escapes the kit")
	ErrTooBig  = errors.New("kit is too large")
	ErrHash    = errors.New("kit hash does not match")
	ErrPass    = errors.New("passphrase does not match")
	ErrAccount = errors.New("account folder is missing")
	ErrFlavor  = errors.New("flavor does not match")
	ErrLocale  = errors.New("locale does not match")
	ErrBuild   = errors.New("build does not match")
	ErrNewer   = errors.New("destination is newer")
	ErrRealm   = errors.New("character realm differs")
)

// File is one payload hash.
type File struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

// Character is a character folder inside the kit.
type Character struct {
	Name  string `json:"name"`
	Realm string `json:"realm"`
}

// Warning is a secret-shaped line. The value is not stored.
type Warning struct {
	File string `json:"file"`
	Line int    `json:"line"`
}

// Manifest is wow-kit v1. AccountID stays inside the kit, not in app config.
type Manifest struct {
	Format      string      `json:"format"`
	Version     int         `json:"version"`
	CreatedAt   string      `json:"created_at"`
	Flavor      string      `json:"flavor"`
	FlavorDir   string      `json:"flavor_dir"`
	Build       string      `json:"build,omitempty"`
	Interface   string      `json:"interface,omitempty"`
	Locale      string      `json:"locale,omitempty"`
	AccountID   string      `json:"account_id"`
	Label       string      `json:"label,omitempty"`
	Scopes      []string    `json:"scopes"`
	Consistency string      `json:"consistency"`
	Files       []File      `json:"files"`
	Omitted     []string    `json:"omitted"`
	Characters  []Character `json:"characters"`
	Warnings    []Warning   `json:"warnings"`
}

// Hit is a kit a search can show. It has no account id.
type Hit struct {
	Path      string `json:"path"`
	Label     string `json:"label,omitempty"`
	Created   string `json:"created,omitempty"`
	Encrypted bool   `json:"encrypted"`
	Size      int64  `json:"size"`
}

// Report is a verified kit. Callers must not print AccountID.
type Report struct {
	Manifest  Manifest
	Encrypted bool
	Label     string
	Verified  bool
}

// RestoreReport counts what was written. Character names are the person's.
type RestoreReport struct {
	Written           int
	SkippedCharacters []string
	Warnings          []Warning
}

func hashOf(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func sealOf(files []File) string {
	h := sha256.New()
	for _, file := range files {
		_, _ = h.Write([]byte(file.Path))
		_, _ = h.Write([]byte{0})
		_, _ = h.Write([]byte(file.SHA256))
		_, _ = h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}

// Seal is the hash a confirm token stores for a verified manifest.
func Seal(m Manifest) string {
	return sealOf(m.Files)
}

func safeSlash(rel string) (string, error) {
	if rel == "" || strings.Contains(rel, `\`) || strings.Contains(rel, ":") {
		return "", ErrEscape
	}
	clean := pathClean(rel)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") || strings.HasPrefix(clean, "/") {
		return "", ErrEscape
	}
	return clean, nil
}

func pathClean(rel string) string {
	return strings.TrimSuffix(filepath.ToSlash(filepath.Clean(rel)), "/")
}

func joinUnder(root, rel string) (string, error) {
	cleanRel, err := safeSlash(rel)
	if err != nil {
		return "", err
	}
	full := filepath.Join(root, filepath.FromSlash(cleanRel))
	rootClean := filepath.Clean(root)
	fullClean := filepath.Clean(full)
	if fullClean != rootClean && !strings.HasPrefix(fullClean, rootClean+string(filepath.Separator)) {
		return "", ErrEscape
	}
	return fullClean, nil
}

func assertDest(dest string) error {
	clean := filepath.Clean(dest)
	if clean == string(filepath.Separator) || clean == "." || !strings.HasSuffix(clean, ".wowkit") {
		return errors.New("kit destination must end in .wowkit")
	}
	return nil
}

func writeAtomic(path string, body []byte, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	tmp := path + ".bornal-tmp"
	if err := os.WriteFile(tmp, body, mode); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func oneLine(value string) string {
	value = strings.ReplaceAll(value, "\r", " ")
	value = strings.ReplaceAll(value, "\n", " ")
	return strings.TrimSpace(value)
}

// FileLabel is the visible kit name. A blank label becomes flavor plus date.
func FileLabel(label, flavor string, now time.Time) string {
	label = oneLine(label)
	if label == "" {
		name := strings.Trim(flavor, "_")
		if name == "" || name == "." {
			name = "wow"
		}
		label = name + " " + now.UTC().Format("2006-01-02")
	}
	return label
}

// FileName makes a single path segment from a label.
func FileName(label string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range label {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		default:
			if !lastDash && b.Len() > 0 {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	name := strings.Trim(b.String(), "-")
	if name == "" {
		name = "kit"
	}
	return name + ".wowkit"
}

func created(m Manifest) time.Time {
	parsed, err := time.Parse(time.RFC3339, m.CreatedAt)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func peekManifest(dir string) (Manifest, error) {
	body, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return Manifest{}, err
	}
	var m Manifest
	if err := json.Unmarshal(body, &m); err != nil {
		return Manifest{}, err
	}
	if m.Format != Format || m.Version != Version {
		return Manifest{}, errors.New("kit format")
	}
	return m, nil
}

func readManifest(dir string) (Manifest, error) {
	body, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return Manifest{}, err
	}
	var m Manifest
	if err := json.Unmarshal(body, &m); err != nil {
		return Manifest{}, err
	}
	if m.Format != Format || m.Version != Version {
		return Manifest{}, errors.New("kit format")
	}
	if err := verifyTree(dir, m); err != nil {
		return Manifest{}, err
	}
	return m, nil
}

func verifyTree(dir string, m Manifest) error {
	want := map[string]File{}
	for _, file := range m.Files {
		rel, err := safeSlash(file.Path)
		if err != nil {
			return err
		}
		if _, ok := want[rel]; ok {
			return ErrHash
		}
		want[rel] = file
	}
	seen := map[string]bool{}
	err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return ErrEscape
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == "." {
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		if rel == "manifest.json" {
			return nil
		}
		seen[rel] = true
		file, ok := want[rel]
		if !ok {
			return ErrHash
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if int64(len(body)) != file.Size || hashOf(body) != file.SHA256 {
			return ErrHash
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(seen) != len(want) {
		return ErrHash
	}
	return nil
}
