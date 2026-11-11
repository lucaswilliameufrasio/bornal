// Package macro reads and plans account macros without writing them.
package macro

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	nameLimit = 16
	bodyLimit = 255
	icon      = "134400"
)

// Macro is one account macro.
type Macro struct {
	ID   string
	Name string
	Icon string
	Body string
}

// Action is what a planned macro would do to the file.
type Action string

const (
	Create   Action = "create"
	Replace  Action = "replace"
	Skip     Action = "skip"
	Conflict Action = "conflict"
)

// Item is one row of a preview.
type Item struct {
	Name   string `json:"name"`
	Action Action `json:"action"`
}

// Plan is the difference between the file and the built-in set.
type Plan struct {
	Items    []Item `json:"items"`
	SHA256   string `json:"sha256"`
	Conflict bool   `json:"conflict"`
	LocaleOK bool   `json:"locale_ok"`
}

var (
	headerLine = regexp.MustCompile(`^VER\s+\d+`)
	macroLine  = regexp.MustCompile(`^(\d{16}) "([^"]*)" "([^"]*)"$`)
	attackLine = regexp.MustCompile(`(?i)^/cast\b(.*)$`)
)

// English reports whether the client locale uses the English spell names.
func English(locale string) bool {
	switch strings.ToLower(strings.TrimSpace(locale)) {
	case "enus", "engb", "en":
		return true
	default:
		return false
	}
}

// Read parses macros-cache.txt. A missing file is an empty list.
func Read(path string) ([]Macro, string, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, hashOf(nil), nil
		}
		return nil, "", err
	}
	macros, err := Parse(body)
	if err != nil {
		return nil, "", err
	}
	return macros, hashOf(body), nil
}

// Parse accepts LF or CRLF. The VER line is a header, then each macro ends at END.
func Parse(body []byte) ([]Macro, error) {
	text := strings.ReplaceAll(string(body), "\r\n", "\n")
	text = strings.TrimPrefix(text, "\ufeff")
	scanner := bufio.NewScanner(strings.NewReader(text))
	macros := make([]Macro, 0)
	seenHeader := false
	var current *Macro
	var lines []string
	for scanner.Scan() {
		line := scanner.Text()
		if current == nil {
			if strings.TrimSpace(line) == "" {
				continue
			}
			if headerLine.MatchString(line) {
				seenHeader = true
				continue
			}
			match := macroLine.FindStringSubmatch(line)
			if match == nil {
				return nil, fmt.Errorf("linha de macro inválida")
			}
			seenHeader = true
			current = &Macro{ID: match[1], Name: match[2], Icon: match[3]}
			lines = nil
			continue
		}
		if line == "END" {
			current.Body = strings.Join(lines, "\n")
			macros = append(macros, *current)
			current = nil
			continue
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if current != nil {
		return nil, errors.New("macro sem END")
	}
	if len(body) > 0 && !seenHeader && len(macros) == 0 {
		return nil, errors.New("arquivo sem VER")
	}
	return macros, nil
}

// PlanSet compares the file with the built-in English set.
func PlanSet(current []Macro, locale string, replace bool) (Plan, []Macro) {
	plan := Plan{Items: make([]Item, 0), LocaleOK: English(locale)}
	next := append([]Macro(nil), current...)
	used := map[string]bool{}
	for _, one := range current {
		used[one.ID] = true
	}
	for _, want := range Catalog() {
		idx := findName(next, want.Name)
		if idx < 0 {
			want.ID = nextID(used)
			used[want.ID] = true
			next = append(next, want)
			plan.Items = append(plan.Items, Item{Name: want.Name, Action: Create})
			continue
		}
		if next[idx].Body == want.Body {
			plan.Items = append(plan.Items, Item{Name: want.Name, Action: Skip})
			continue
		}
		if !replace {
			plan.Conflict = true
			plan.Items = append(plan.Items, Item{Name: want.Name, Action: Conflict})
			continue
		}
		next[idx].Body = want.Body
		next[idx].Icon = want.Icon
		plan.Items = append(plan.Items, Item{Name: want.Name, Action: Replace})
	}
	return plan, next
}

// Format writes the cache with CRLF, which is what the client reads.
func Format(macros []Macro) ([]byte, error) {
	var buf bytes.Buffer
	if _, err := buf.WriteString("VER 3\r\n"); err != nil {
		return nil, err
	}
	for _, one := range macros {
		if err := validate(one); err != nil {
			return nil, err
		}
		if _, err := fmt.Fprintf(&buf, "%s \"%s\" \"%s\"\r\n", one.ID, one.Name, one.Icon); err != nil {
			return nil, err
		}
		body := strings.ReplaceAll(one.Body, "\r\n", "\n")
		for _, line := range strings.Split(body, "\n") {
			if _, err := fmt.Fprintf(&buf, "%s\r\n", line); err != nil {
				return nil, err
			}
		}
		if _, err := buf.WriteString("END\r\n"); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

// WriteFile replaces path only after the bytes are complete.
func WriteFile(path string, macros []Macro) error {
	body, err := Format(macros)
	if err != nil {
		return err
	}
	temp := path + ".bornal-tmp"
	if err := os.WriteFile(temp, body, 0o644); err != nil {
		return err
	}
	return os.Rename(temp, path)
}

// Backup copies the current file beside it. A missing source writes nothing.
func Backup(path string) error {
	src, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer src.Close()
	dst, err := os.OpenFile(path+".bornal-bak", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer dst.Close()
	if _, err := io.Copy(dst, src); err != nil {
		return err
	}
	return dst.Close()
}

func validate(one Macro) error {
	if len(one.Name) == 0 || len(one.Name) > nameLimit {
		return fmt.Errorf("nome de macro inválido")
	}
	if strings.ContainsAny(one.Name, "\"\r\n") {
		return fmt.Errorf("nome de macro inválido")
	}
	if len(one.Body) > bodyLimit {
		return fmt.Errorf("corpo de macro longo demais")
	}
	for _, line := range strings.Split(one.Body, "\n") {
		if castsAttack(line) {
			return fmt.Errorf("/cast Attack alterna o ataque e não pode entrar na macro")
		}
	}
	return nil
}

func castsAttack(line string) bool {
	match := attackLine.FindStringSubmatch(strings.TrimSpace(line))
	if match == nil {
		return false
	}
	rest := match[1]
	if i := strings.LastIndex(rest, ";"); i >= 0 {
		rest = rest[i+1:]
	}
	rest = strings.Trim(strings.TrimSpace(rest), "[]")
	fields := strings.Fields(rest)
	if len(fields) == 0 {
		return false
	}
	return strings.EqualFold(fields[len(fields)-1], "Attack")
}

func findName(macros []Macro, name string) int {
	for i, one := range macros {
		if one.Name == name {
			return i
		}
	}
	return -1
}

func nextID(used map[string]bool) string {
	var max uint64
	for id := range used {
		n, err := strconv.ParseUint(id, 10, 64)
		if err == nil && n > max {
			max = n
		}
	}
	for {
		max++
		id := fmt.Sprintf("%016d", max)
		if !used[id] {
			return id
		}
	}
}

func hashOf(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

// Catalog is the English warrior set. Spell names are not translated here.
func Catalog() []Macro {
	return []Macro{
		spell("Rend", "#showtooltip Rend\n/startattack\n/cast Rend"),
		spell("Charge", "#showtooltip Charge\n/startattack\n/cast Charge"),
		spell("Hamstring", "#showtooltip Hamstring\n/startattack\n/cast Hamstring"),
		spell("Revenge", "#showtooltip Revenge\n/startattack\n/cast [nostance:2] Defensive Stance; Revenge"),
		spell("Taunt", "#showtooltip Taunt\n/startattack\n/cast [nostance:2] Defensive Stance; [@mouseover,harm,nodead][] Taunt"),
		spell("Battle Shout", "#showtooltip Battle Shout\n/cast Bloodrage\n/cast Battle Shout"),
		spell("Victory Rush", "#showtooltip Victory Rush\n/startattack\n/cast Victory Rush"),
		spell("Shield Slam", "#showtooltip Shield Slam\n/startattack\n/cast Shield Slam"),
		spell("Overpower", "#showtooltip Overpower\n/startattack\n/cast [nostance:1] Battle Stance; Overpower"),
		spell("Whirlwind", "#showtooltip Whirlwind\n/startattack\n/cast [nostance:3] Berserker Stance; Whirlwind"),
	}
}

func spell(name, body string) Macro {
	return Macro{Name: name, Icon: icon, Body: body}
}
