package envyaml // import "go.tmthrgd.dev/env-yaml"

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

var testEnvDir string

func Init() {
	env, err := Load()
	if err != nil {
		panic(err)
	}

	for k, v := range env {
		if _, dup := os.LookupEnv(k); dup {
			continue
		}

		os.Setenv(k, v)
	}
}

func Load() (map[string]string, error) {
	data, err := ioutil.ReadFile(testEnvDir + ".env.yaml")
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, err
	}

	var env map[string]string
	err = yaml.Unmarshal(data, &env)

	for k := range env {
		if !strings.ContainsAny(k, "=$%") &&
			strings.IndexFunc(k, isNotPrint) < 0 {
			continue
		}

		delete(env, k)

		if err == nil {
			err = &InvalidKeyError{k}
		}
	}

	return env, err
}

func ShellEscaped() ([]string, error) {
	env, err := Load()

	kv := make([]string, 0, len(env))
	for k, v := range env {
		v, quoteErr := quoteShell(v)
		if quoteErr == nil {
			kv = append(kv, k+"="+v)
		} else if err == nil {
			err = quoteErr
		}
	}

	sort.Strings(kv)
	return kv, err
}

func quoteShell(v string) (string, error) {
	const special = "\\'\"`${[|&;<>()*?!\r\n"
	switch {
	case strings.IndexFunc(v, isNotPrint) >= 0:
		return "", &InvalidValueError{v}
	case !strings.HasPrefix(v, "~") && !strings.ContainsAny(v, special):
		return v, nil
	case !strings.ContainsAny(v, "'\r\n"):
		return "'" + v + "'", nil
	}

	var s strings.Builder
	s.WriteByte('"')

	for _, r := range v {
		switch {
		case r == '\r':
			s.WriteString(`\r`)
		case r == '\n':
			s.WriteString(`\n`)
		case strings.ContainsRune(special, r):
			s.WriteByte('\\')
			s.WriteRune(r)
		default:
			s.WriteRune(r)
		}
	}

	s.WriteByte('"')
	return s.String(), nil
}

func isNotPrint(r rune) bool {
	const allowed = "\t\r\n"
	return !strconv.IsPrint(r) && !strings.ContainsRune(allowed, r)
}

type InvalidKeyError struct{ Key string }

func (e *InvalidKeyError) Error() string {
	if idx := strings.IndexFunc(e.Key, isNotPrint); idx > 0 {
		return fmt.Sprintf("env-yaml: key (%q) contains unprintable character %U", e.Key, e.Key[idx])
	}

	return fmt.Sprintf("env-yaml: invalid key %q", e.Key)
}

type InvalidValueError struct{ Value string }

func (e *InvalidValueError) Error() string {
	if idx := strings.IndexFunc(e.Value, isNotPrint); idx > 0 {
		return fmt.Sprintf("env-yaml: value (%q) contains unprintable character %U", e.Value, e.Value[idx])
	}

	return fmt.Sprintf("env-yaml: invalid value %q", e.Value)
}
