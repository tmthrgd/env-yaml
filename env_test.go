package envyaml

import (
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func restoreEnv() func() {
	env := os.Environ()
	return func() {
		os.Clearenv()

		for _, kv := range env {
			var k, v string
			if idx := strings.Index(kv, "="); idx >= 0 {
				k, v = kv[:idx], kv[idx+1:]
			} else {
				k = kv
			}

			os.Setenv(k, v)
		}
	}
}

func TestLoad(t *testing.T) {
	defer func(old string) { testEnvDir = old }(testEnvDir)
	testEnvDir = "testdata/"

	env, err := Load()
	require.NoError(t, err)
	assert.Equal(t, map[string]string{
		"DUPLICATE": "NEW",
		"KEY":       "VALUE",
	}, env)
}

func TestInit(t *testing.T) {
	defer restoreEnv()()
	defer func(old string) { testEnvDir = old }(testEnvDir)
	testEnvDir = "testdata/"

	os.Clearenv()
	assert.NoError(t, os.Setenv("DUPLICATE", "OLD"))

	assert.NotPanics(t, Init)

	env := os.Environ()
	sort.Strings(env)
	assert.Equal(t, []string{
		"DUPLICATE=OLD",
		"KEY=VALUE",
	}, env)
}

func TestLoadInvalid(t *testing.T) {
	defer func(old string) { testEnvDir = old }(testEnvDir)

	for _, dir := range []string{
		"testdata/invalid-yaml/",
		"testdata/invalid-unprintable-key/",
	} {
		testEnvDir = dir

		env, err := Load()
		assert.Error(t, err, "dir=%s", dir)
		assert.Empty(t, env, "dir=%s", dir)
	}

	testEnvDir = "testdata/invalid-key/"

	env, err := Load()
	assert.Error(t, err)
	assert.Equal(t, map[string]string{
		"VALID": "VALUE",
	}, env)
}

func TestInitInvalid(t *testing.T) {
	defer restoreEnv()()
	defer func(old string) { testEnvDir = old }(testEnvDir)

	for _, dir := range []string{
		"testdata/invalid-yaml/",
		"testdata/invalid-key/",
		"testdata/invalid-unprintable-key/",
	} {
		testEnvDir = dir

		assert.Panics(t, Init)
	}
}

func TestShellEscaped(t *testing.T) {
	defer func(old string) { testEnvDir = old }(testEnvDir)
	testEnvDir = "testdata/needs-quoting/"

	env, err := ShellEscaped()
	require.NoError(t, err)
	assert.Equal(t, []string{
		"DOUBLEESCAPED=\"VALUE \\$NEEDS \\$\\(ESCAPING\\) \\'to\\' \\\"be\\\" \\`valid\\` \\\\and \\[safe] \\<to\\> \\(use\\)\\!\\?\"",
		"KEY=VALUE",
		`NEWLINE="VALUE\nVALUE"`,
		"SINGLEESCAPED='VALUE $NEEDS $(ESCAPING) to \"be\" `valid` \\and [safe] <to> (use)!?'",
		"UNICODE=VALUE\u2620",
	}, env)
}

func TestShellEscapedInvalid(t *testing.T) {
	defer func(old string) { testEnvDir = old }(testEnvDir)
	testEnvDir = "testdata/invalid-unprintable-value/"

	env, err := ShellEscaped()
	assert.Error(t, err)
	assert.Equal(t, []string{
		"KEY=VALUE",
	}, env)
}

func TestQuoteShell(t *testing.T) {
	for v, q := range map[string]string{
		"VALUE": "VALUE",
		"VALUE $NEEDS $(ESCAPING) to \"be\" `valid` \\and [safe] <to> (use)!?":   "'VALUE $NEEDS $(ESCAPING) to \"be\" `valid` \\and [safe] <to> (use)!?'",
		"VALUE $NEEDS $(ESCAPING) 'to' \"be\" `valid` \\and [safe] <to> (use)!?": "\"VALUE \\$NEEDS \\$\\(ESCAPING\\) \\'to\\' \\\"be\\\" \\`valid\\` \\\\and \\[safe] \\<to\\> \\(use\\)\\!\\?\"",
		"VALUE\nVALUE": `"VALUE\nVALUE"`,
		"VALUE\tVALUE": "VALUE\tVALUE",
		"~VALUE":       "'~VALUE'",
		"VALUE~":       "VALUE~",
		"VALUE\u2620":  "VALUE\u2620",
	} {
		got, err := quoteShell(v)
		if assert.NoErrorf(t, err, "quoteShell(%q)", v) {
			assert.Equalf(t, q, got, "quoteShell(%q)", v)
		}
	}

	for _, v := range []string{
		"VALUE\u200B",
	} {
		_, err := quoteShell(v)
		assert.Errorf(t, err, "quoteShell(%q)", v)
	}
}
