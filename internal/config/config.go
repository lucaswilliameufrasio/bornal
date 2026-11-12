// Package config stores the answers from the first-run questions.
package config

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// File is the saved setup. Language is "pt-BR" or "en".
type File struct {
	Language    string `json:"language,omitempty"`
	InstallPath string `json:"install_path,omitempty"`
}

// Path is the config file for this user.
func Path(getenv func(string) string) (string, error) {
	if getenv == nil {
		return "", errors.New("ambiente ausente")
	}
	if dir := getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, "bornal", "config.json"), nil
	}
	home := getenv("HOME")
	if home == "" {
		home = getenv("USERPROFILE")
	}
	if home == "" {
		return "", errors.New("diretório do usuário ausente")
	}
	return filepath.Join(home, ".config", "bornal", "config.json"), nil
}

// Load returns an empty file when the person has not answered yet.
func Load(getenv func(string) string) (File, error) {
	path, err := Path(getenv)
	if err != nil {
		return File{}, nil
	}
	body, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return File{}, nil
		}
		return File{}, err
	}
	var file File
	if err := json.Unmarshal(body, &file); err != nil {
		return File{}, err
	}
	return file, nil
}

// Save writes the file so a crash cannot leave a half-written config.
func Save(getenv func(string) string, file File) error {
	path, err := Path(getenv)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	body, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	temp := path + ".tmp"
	if err := os.WriteFile(temp, body, 0o600); err != nil {
		return err
	}
	return os.Rename(temp, path)
}
