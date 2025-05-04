package rules

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMissingIfNotExists(t *testing.T) {
	t.Parallel()

	t.Run("Should find violation for relation", func(t *testing.T) {
		t.Parallel()

		tree := mustParse(t, "CREATE TABLE pgcheck (id integer PRIMARY KEY);")
		require.Len(t, tree.Stmts, 1)

		res, err := missingIfNotExists(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		require.Len(t, res, 1)

		assert.EqualValues(t, 0, res[0].StmtStart)
		assert.EqualValues(t, 45, res[0].StmtEnd)
		assert.Equal(t, testCode, res[0].Code)
		assert.Equal(t, testSlug, res[0].Slug)
		assert.Equal(t, testHelp, res[0].Help)
	})

	t.Run("Should find violation for index", func(t *testing.T) {
		t.Parallel()

		tree := mustParse(t, "CREATE INDEX pgcheck_key ON pgcheck(id);")
		require.Len(t, tree.Stmts, 1)

		res, err := missingIfNotExists(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		require.Len(t, res, 1)

		assert.EqualValues(t, 0, res[0].StmtStart)
		assert.EqualValues(t, 39, res[0].StmtEnd)
		assert.Equal(t, testCode, res[0].Code)
		assert.Equal(t, testSlug, res[0].Slug)
		assert.Equal(t, testHelp, res[0].Help)
	})

	t.Run("Should find multiple violations", func(t *testing.T) {
		t.Parallel()

		var b strings.Builder
		b.WriteString("CREATE TABLE pgcheck (id integer PRIMARY KEY);\n")
		b.WriteString("ALTER TABLE pgcheck RENAME COLUMN value TO new_value;\n")
		b.WriteString("CREATE INDEX pgcheck_key ON pgcheck(id);\n")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 3)

		res, err := missingIfNotExists(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		require.Len(t, res, 2)

		assert.EqualValues(t, 0, res[0].StmtStart)
		assert.EqualValues(t, 100, res[1].StmtStart)
	})

	t.Run("Should not return statements that are safe", func(t *testing.T) {
		t.Parallel()

		var b strings.Builder
		b.WriteString("CREATE TABLE IF NOT EXISTS pgcheck (id integer PRIMARY KEY);")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 1)

		res, err := missingIfNotExists(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		assert.Empty(t, res)
	})

	t.Run("Should find no violations", func(t *testing.T) {
		t.Parallel()

		var b strings.Builder
		b.WriteString("ALTER TABLE pgcheck RENAME COLUMN value TO new_value;\n")
		b.WriteString("ALTER TABLE pgcheck RENAME COLUMN value TO new_value;\n")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 2)

		res, err := missingIfNotExists(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		assert.Empty(t, res)
	})
}
