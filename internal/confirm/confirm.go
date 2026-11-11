// Package confirm stores a one-time preview token for one side.
// The CLI, the MCP server, and the console do not accept each other's tokens.
package confirm

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"time"
	"unicode"
)

const lifetime = 15 * time.Minute

// Issue writes a token that only the same side can take.
func Issue(dir, side string, body []byte) (string, error) {
	if err := os.MkdirAll(filepath.Join(dir, side), 0o700); err != nil {
		return "", err
	}
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", err
	}
	token := hex.EncodeToString(raw[:])
	path := filepath.Join(dir, side, token)
	if err := os.WriteFile(path, body, 0o600); err != nil {
		return "", err
	}
	return token, nil
}

// Take returns the body and removes the token. A token from another side misses.
func Take(dir, side, token string) ([]byte, error) {
	if !hexToken(token) {
		return nil, os.ErrNotExist
	}
	path := filepath.Join(dir, side, token)
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if time.Since(info.ModTime()) > lifetime {
		_ = os.Remove(path)
		return nil, os.ErrNotExist
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := os.Remove(path); err != nil {
		return nil, err
	}
	return body, nil
}

// Dir is the state directory for tokens and the last kit export.
func Dir(getenv func(string) string) (string, error) {
	if getenv == nil {
		return "", errors.New("ambiente ausente")
	}
	if dir := getenv("XDG_STATE_HOME"); dir != "" {
		return filepath.Join(dir, "bornal"), nil
	}
	home := getenv("HOME")
	if home == "" {
		home = getenv("USERPROFILE")
	}
	if home == "" {
		return "", errors.New("diretório do usuário ausente")
	}
	return filepath.Join(home, ".local", "state", "bornal"), nil
}

func hexToken(token string) bool {
	if len(token) != 32 {
		return false
	}
	for _, r := range token {
		if !unicode.Is(unicode.ASCII_Hex_Digit, r) {
			return false
		}
	}
	return true
}
