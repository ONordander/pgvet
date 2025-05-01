package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLint(t *testing.T) {
	t.Parallel()

	shouldWriteTestdata := os.Getenv("OVERWRITE_TESTDATA") == "true"

	cases := map[string]struct {
		file         string
		expectedfile string
		configfile   *string
	}{
		"breaking":    {"testdata/breaking.sql", "testdata/breaking.out", nil},
		"nullability": {"testdata/nullability.sql", "testdata/nullability.out", nil},
		"idempotency": {"testdata/idempotency.sql", "testdata/idempotency.out", nil},
		"locking":     {"testdata/locking.sql", "testdata/locking.out", nil},
		"formatting":  {"testdata/formatting.sql", "testdata/formatting.out", nil},
		"noerrors":    {"testdata/noerrors.sql", "testdata/noerrors.out", nil},
		"with-config": {"testdata/with-config.sql", "testdata/with-config.out", ptr("testdata/with-config.yaml")},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var wOut, wErr strings.Builder
			rc := lint(&wOut, &wErr, false, []string{tc.file}, tc.configfile, formatText)
			require.Zero(t, rc, wErr.String())

			if shouldWriteTestdata {
				mustWriteFile(t, wOut.String(), tc.expectedfile)
				return
			}

			expected := mustReadFile(t, tc.expectedfile)
			assert.Equal(t, expected, wOut.String())
			assert.Empty(t, wErr.String())
		})
	}
}

func TestLintPatterns(t *testing.T) {
	t.Parallel()

	t.Run("Folder", func(t *testing.T) {
		t.Parallel()
		var wOut, wErr strings.Builder
		rc := lint(&wOut, &wErr, false, []string{"testdata/patterns/*"}, nil, formatText)
		require.Zero(t, rc, wErr.String())

		out := wOut.String()
		assert.Contains(t, out, "1-pattern.sql")
		assert.Contains(t, out, "2-pattern.sql")
		assert.Contains(t, out, "1.sql")
	})

	t.Run("Pattern", func(t *testing.T) {
		t.Parallel()
		var wOut, wErr strings.Builder
		rc := lint(&wOut, &wErr, false, []string{"testdata/patterns/*-pattern.sql"}, nil, formatText)
		require.Zero(t, rc, wErr.String())

		out := wOut.String()
		assert.Contains(t, out, "1-pattern.sql")
		assert.Contains(t, out, "2-pattern.sql")
		assert.NotContains(t, out, "1.sql")
	})

	t.Run("Multiple patterns", func(t *testing.T) {
		t.Parallel()
		var wOut, wErr strings.Builder
		rc := lint(&wOut, &wErr, false, []string{"testdata/**/1-pattern.sql", "testdata/**/2-pattern.sql"}, nil, formatText)
		require.Zero(t, rc, wErr.String())

		out := wOut.String()
		assert.Contains(t, out, "1-pattern.sql")
		assert.Contains(t, out, "2-pattern.sql")
		assert.NotContains(t, out, "1.sql")
	})
}

func TestLintFormatJson(t *testing.T) {
	t.Parallel()

	var wOut, wErr strings.Builder
	rc := lint(&wOut, &wErr, false, []string{"testdata/patterns/*"}, nil, formatJson)
	require.Zero(t, rc, wErr.String())

	out := wOut.String()
	var report Report
	err := json.NewDecoder(bytes.NewBuffer([]byte(out))).Decode(&report)
	require.NoError(t, err)

	assert.NotEmpty(t, report)
}

func TestLintError(t *testing.T) {
	t.Parallel()

	t.Run("Syntax error", func(t *testing.T) {
		t.Parallel()
		var wOut, wErr strings.Builder
		rc := lint(&wOut, &wErr, false, []string{"testdata/error.sql"}, nil, formatText)
		require.NotZero(t, rc)

		assert.Empty(t, wOut.String())
		assert.Contains(t, wErr.String(), "Failed to parse file")
	})

	t.Run("No files", func(t *testing.T) {
		t.Parallel()
		var wOut, wErr strings.Builder
		rc := lint(&wOut, &wErr, false, []string{"testdata/missingfiles*.sql"}, nil, formatText)
		require.NotZero(t, rc)

		assert.Empty(t, wOut.String())
		assert.Contains(t, wErr.String(), "No files found")
	})

	t.Run("Missing config", func(t *testing.T) {
		t.Parallel()
		var wOut, wErr strings.Builder
		rc := lint(&wOut, &wErr, false, []string{"testdata/noerrors.sql"}, ptr("no-config.yaml"), formatText)
		require.NotZero(t, rc)

		assert.Empty(t, wOut.String())
		assert.Contains(t, wErr.String(), "Failed to parse config")
	})
}

func mustReadFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err, "failed to read file: %q", path)
	return string(content)
}

func mustWriteFile(t *testing.T, content, path string) {
	t.Helper()
	err := os.WriteFile(path, []byte(content), 0o664)
	require.NoError(t, err, "failed to write file: %q", path)
}

func ptr(s string) *string {
	return &s
}
