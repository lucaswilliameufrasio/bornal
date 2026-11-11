package cli

import (
	"fmt"
	"strconv"
	"strings"
)

type options struct {
	Path    string
	Flavor  string
	Confirm string
	Label   string
	Out     string
	Account int
	Replace bool
	Encrypt bool
	RealmOK bool
	SawPath bool
	Rest    []string
}

func parseOptions(args []string) (options, error) {
	var opt options
	for i := 0; i < len(args); i++ {
		arg := args[i]
		name, value, glued, ok := splitFlag(arg)
		if !ok {
			opt.Rest = append(opt.Rest, arg)
			continue
		}
		take := func() (string, error) {
			if glued {
				return value, nil
			}
			if i+1 >= len(args) {
				return "", fmt.Errorf("%s", name)
			}
			i++
			return args[i], nil
		}
		switch name {
		case "path":
			got, err := take()
			if err != nil {
				return options{}, err
			}
			opt.Path = got
			opt.SawPath = true
		case "flavor":
			got, err := take()
			if err != nil {
				return options{}, err
			}
			opt.Flavor = got
		case "confirm":
			got, err := take()
			if err != nil {
				return options{}, err
			}
			opt.Confirm = got
		case "label":
			got, err := take()
			if err != nil {
				return options{}, err
			}
			opt.Label = got
		case "out":
			got, err := take()
			if err != nil {
				return options{}, err
			}
			opt.Out = got
		case "account":
			got, err := take()
			if err != nil {
				return options{}, err
			}
			n, convErr := strconv.Atoi(got)
			if convErr != nil || n < 1 {
				return options{}, fmt.Errorf("%s", name)
			}
			opt.Account = n
		case "replace":
			opt.Replace = boolFlag(value, glued)
		case "encrypt":
			opt.Encrypt = boolFlag(value, glued)
		case "realm-ok":
			opt.RealmOK = boolFlag(value, glued)
		default:
			return options{}, fmt.Errorf("%s", arg)
		}
	}
	return opt, nil
}

func splitFlag(arg string) (string, string, bool, bool) {
	if !strings.HasPrefix(arg, "--") || arg == "--" {
		return "", "", false, false
	}
	body := strings.TrimPrefix(arg, "--")
	name, value, glued := strings.Cut(body, "=")
	return name, value, glued, name != ""
}

func boolFlag(value string, glued bool) bool {
	if !glued {
		return true
	}
	switch strings.ToLower(value) {
	case "false", "0", "no":
		return false
	default:
		return true
	}
}
