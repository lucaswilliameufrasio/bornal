package kit

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"filippo.io/age"
)

// CreateOptions selects one account inside one flavor directory.
type CreateOptions struct {
	InstallPath string
	Account     string
	Locale      string
	Build       string
	Dest        string
	Label       string
	Encrypt     bool
	Passphrase  []byte
	Work        int
	Now         time.Time
}

// Create writes a directory kit, or one age file when Encrypt is set.
func Create(opts CreateOptions) (Manifest, error) {
	if err := assertDest(opts.Dest); err != nil {
		return Manifest{}, err
	}
	if strings.TrimSpace(opts.Account) == "" {
		return Manifest{}, ErrAccount
	}
	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}
	now = now.UTC()
	label := FileLabel(opts.Label, filepath.Base(opts.InstallPath), now)
	files, characters, omitted, warnings, interfaceID, err := collect(opts.InstallPath, opts.Account)
	if err != nil {
		return Manifest{}, err
	}
	manifest := Manifest{
		Format:      Format,
		Version:     Version,
		CreatedAt:   now.Format(time.RFC3339),
		Flavor:      filepath.Base(opts.InstallPath),
		FlavorDir:   filepath.Base(opts.InstallPath),
		Build:       opts.Build,
		Interface:   interfaceID,
		Locale:      opts.Locale,
		AccountID:   opts.Account,
		Label:       label,
		Scopes:      scopesOf(files),
		Consistency: "ok",
		Files:       make([]File, 0, len(files)),
		Omitted:     omitted,
		Characters:  characters,
		Warnings:    warnings,
	}
	for _, file := range files {
		manifest.Files = append(manifest.Files, File{
			Path:   file.Rel,
			SHA256: hashOf(file.Data),
			Size:   int64(len(file.Data)),
		})
	}
	body, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return Manifest{}, err
	}
	body = append(body, '\n')
	if !opts.Encrypt {
		if err := publishDir(opts.Dest, files, body); err != nil {
			return Manifest{}, err
		}
		return manifest, nil
	}
	if len(bytes.TrimSpace(opts.Passphrase)) == 0 {
		return Manifest{}, ErrPass
	}
	payload, err := tarOf(files, body)
	if err != nil {
		return Manifest{}, err
	}
	encoded, err := encryptPayload(payload, opts.Passphrase, opts.Work, label, now)
	if err != nil {
		return Manifest{}, err
	}
	if err := publishFile(opts.Dest, encoded); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

func scopesOf(files []stored) []string {
	has := map[string]bool{}
	for _, file := range files {
		switch {
		case strings.HasPrefix(file.Rel, "account/"):
			has["account"] = true
		case strings.HasPrefix(file.Rel, "characters/"):
			has["characters"] = true
		case strings.HasPrefix(file.Rel, "addons/"):
			has["addons"] = true
		case strings.HasPrefix(file.Rel, "config/"):
			has["config"] = true
		}
	}
	order := []string{"account", "characters", "addons", "config"}
	out := make([]string, 0, len(order))
	for _, name := range order {
		if has[name] {
			out = append(out, name)
		}
	}
	return out
}

func publishDir(dest string, files []stored, manifest []byte) error {
	tmp := dest + ".bornal-tmp"
	if err := os.RemoveAll(tmp); err != nil {
		return err
	}
	if err := os.MkdirAll(tmp, 0o700); err != nil {
		return err
	}
	for _, file := range files {
		target, err := joinUnder(tmp, file.Rel)
		if err != nil {
			_ = os.RemoveAll(tmp)
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			_ = os.RemoveAll(tmp)
			return err
		}
		if err := os.WriteFile(target, file.Data, 0o600); err != nil {
			_ = os.RemoveAll(tmp)
			return err
		}
	}
	if err := os.WriteFile(filepath.Join(tmp, "manifest.json"), manifest, 0o600); err != nil {
		_ = os.RemoveAll(tmp)
		return err
	}
	if err := os.RemoveAll(dest); err != nil {
		_ = os.RemoveAll(tmp)
		return err
	}
	if err := os.Rename(tmp, dest); err != nil {
		_ = os.RemoveAll(tmp)
		return err
	}
	return nil
}

func publishFile(dest string, body []byte) error {
	info, err := os.Lstat(dest)
	if err == nil && info.IsDir() {
		return errors.New("kit destination is a directory")
	}
	tmp := dest + ".bornal-tmp"
	if err := os.MkdirAll(filepath.Dir(dest), 0o700); err != nil {
		return err
	}
	if err := os.WriteFile(tmp, body, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, dest)
}

func encryptPayload(payload, passphrase []byte, work int, label string, now time.Time) ([]byte, error) {
	recipient, err := age.NewScryptRecipient(string(passphrase))
	if err != nil {
		return nil, ErrPass
	}
	if work > 0 {
		recipient.SetWorkFactor(work)
	}
	var buf bytes.Buffer
	if _, err := fmt.Fprintf(&buf, "WOWKIT1\ndate: %s\nlabel: %s\nkdf: age-scrypt\n\n", now.Format(time.RFC3339), oneLine(label)); err != nil {
		return nil, err
	}
	writer, err := age.Encrypt(&buf, recipient)
	if err != nil {
		return nil, err
	}
	if _, err := writer.Write(payload); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decryptPayload(body, passphrase []byte) ([]byte, string, string, error) {
	label, date, payload, err := splitHeader(body)
	if err != nil {
		return nil, "", "", err
	}
	if len(bytes.TrimSpace(passphrase)) == 0 {
		return nil, label, date, ErrPass
	}
	identity, err := age.NewScryptIdentity(string(passphrase))
	if err != nil {
		return nil, label, date, ErrPass
	}
	reader, err := age.Decrypt(bytes.NewReader(payload), identity)
	if err != nil {
		return nil, label, date, ErrPass
	}
	plain, err := io.ReadAll(io.LimitReader(reader, maxTotal+1))
	if err != nil {
		return nil, label, date, ErrPass
	}
	if len(plain) > maxTotal {
		return nil, label, date, ErrTooBig
	}
	return plain, label, date, nil
}

func splitHeader(body []byte) (string, string, []byte, error) {
	if !bytes.HasPrefix(body, []byte("WOWKIT1\n")) {
		return "", "", nil, errors.New("not an encrypted kit")
	}
	sep := bytes.Index(body, []byte("\n\n"))
	if sep < 0 || sep > 4096 {
		return "", "", nil, errors.New("kit header")
	}
	header := string(body[:sep])
	var label, date, kdf string
	for _, line := range strings.Split(header, "\n") {
		key, value, ok := strings.Cut(line, ": ")
		if !ok {
			continue
		}
		switch key {
		case "date":
			date = value
		case "label":
			label = value
		case "kdf":
			kdf = value
		}
	}
	if kdf != "age-scrypt" {
		return "", "", nil, errors.New("kit kdf")
	}
	return label, date, body[sep+2:], nil
}

// Header reads the clear WOWKIT1 lines. It does not decrypt.
func Header(path string) (label, date string, encrypted bool, err error) {
	info, statErr := os.Lstat(path)
	if statErr != nil {
		return "", "", false, statErr
	}
	if info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return "", "", false, nil
	}
	file, err := os.Open(path)
	if err != nil {
		return "", "", false, err
	}
	defer file.Close()
	buf := make([]byte, 4096)
	n, readErr := file.Read(buf)
	if readErr != nil && readErr != io.EOF {
		return "", "", false, readErr
	}
	if !bytes.HasPrefix(buf[:n], []byte("WOWKIT1\n")) {
		return "", "", false, nil
	}
	label, date, _, err = splitHeader(buf[:n])
	if err != nil && !bytes.Contains(buf[:n], []byte("\n\n")) {
		return "", "", true, nil
	}
	return label, date, true, nil
}
