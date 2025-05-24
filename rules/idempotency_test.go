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

		tree := mustParse(t, "CREATE TABLE pgvet (id integer PRIMARY KEY);")
		require.Len(t, tree.Stmts, 1)

		res, err := missingIfNotExists(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		require.Len(t, res, 1)

		assert.EqualValues(t, 0, res[0].StmtStart)
		assert.EqualValues(t, 43, res[0].StmtEnd)
		assert.Equal(t, testCode, res[0].Code)
		assert.Equal(t, testSlug, res[0].Slug)
		assert.Equal(t, testHelp, res[0].Help)
	})

	t.Run("Should find violation for index", func(t *testing.T) {
		t.Parallel()

		tree := mustParse(t, "CREATE INDEX pgvet_key ON pgvet(id);")
		require.Len(t, tree.Stmts, 1)

		res, err := missingIfNotExists(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		require.Len(t, res, 1)

		assert.EqualValues(t, 0, res[0].StmtStart)
		assert.EqualValues(t, 35, res[0].StmtEnd)
		assert.Equal(t, testCode, res[0].Code)
		assert.Equal(t, testSlug, res[0].Slug)
		assert.Equal(t, testHelp, res[0].Help)
	})

	t.Run("Should find violation for add column", func(t *testing.T) {
		t.Parallel()

		tree := mustParse(t, "ALTER TABLE pgvet ADD COLUMN value text;")
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
		b.WriteString("CREATE TABLE pgvet (id integer PRIMARY KEY);\n")
		b.WriteString("ALTER TABLE pgvet ADD COLUMN value text;\n")
		b.WriteString("CREATE INDEX pgvet_key ON pgvet(id);\n")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 3)

		res, err := missingIfNotExists(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		require.Len(t, res, 3)

		assert.EqualValues(t, 0, res[0].StmtStart)
		assert.EqualValues(t, 44, res[1].StmtStart)
		assert.EqualValues(t, 85, res[2].StmtStart)
	})

	t.Run("Should not return statements that are safe", func(t *testing.T) {
		t.Parallel()

		var b strings.Builder
		b.WriteString("CREATE TABLE IF NOT EXISTS pgvet (id integer PRIMARY KEY);\n")
		b.WriteString("ALTER TABLE pgvet ADD COLUMN IF NOT EXISTS value text;\n")
		b.WriteString("CREATE INDEX IF NOT EXISTS pgvet_key ON pgvet(id);\n")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 3)

		res, err := missingIfNotExists(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		assert.Empty(t, res)
	})

	t.Run("Should find no violations", func(t *testing.T) {
		t.Parallel()

		var b strings.Builder
		b.WriteString("ALTER TABLE pgvet RENAME COLUMN value TO new_value;\n")
		b.WriteString("ALTER TABLE pgvet RENAME COLUMN value TO new_value;\n")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 2)

		res, err := missingIfNotExists(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		assert.Empty(t, res)
	})
}

func TestMissingIfExists(t *testing.T) {
	t.Parallel()

	t.Run("Should find violation for table", func(t *testing.T) {
		t.Parallel()

		tree := mustParse(t, "DROP TABLE pgvet;")
		require.Len(t, tree.Stmts, 1)

		res, err := missingIfExists(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		require.Len(t, res, 1)

		assert.EqualValues(t, 0, res[0].StmtStart)
		assert.EqualValues(t, 16, res[0].StmtEnd)
		assert.Equal(t, testCode, res[0].Code)
		assert.Equal(t, testSlug, res[0].Slug)
		assert.Equal(t, testHelp, res[0].Help)
	})

	t.Run("Should find violation for index", func(t *testing.T) {
		t.Parallel()

		tree := mustParse(t, "DROP INDEX pgvet_idx;")
		require.Len(t, tree.Stmts, 1)

		res, err := missingIfExists(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		require.Len(t, res, 1)

		assert.EqualValues(t, 0, res[0].StmtStart)
		assert.EqualValues(t, 20, res[0].StmtEnd)
		assert.Equal(t, testCode, res[0].Code)
		assert.Equal(t, testSlug, res[0].Slug)
		assert.Equal(t, testHelp, res[0].Help)
	})

	t.Run("Should find violation for column", func(t *testing.T) {
		t.Parallel()

		tree := mustParse(t, "ALTER TABLE pgvet DROP COLUMN name;")
		require.Len(t, tree.Stmts, 1)

		res, err := missingIfExists(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		require.Len(t, res, 1)

		assert.EqualValues(t, 0, res[0].StmtStart)
		assert.EqualValues(t, 34, res[0].StmtEnd)
		assert.Equal(t, testCode, res[0].Code)
		assert.Equal(t, testSlug, res[0].Slug)
		assert.Equal(t, testHelp, res[0].Help)
	})

	t.Run("Should find multiple violations", func(t *testing.T) {
		t.Parallel()

		var b strings.Builder
		b.WriteString("DROP TABLE pgvet;\n")
		b.WriteString("ALTER TABLE pgvet ADD COLUMN value text;\n")
		b.WriteString("DROP INDEX pgvet_key;\n")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 3)

		res, err := missingIfExists(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		require.Len(t, res, 2)

		assert.EqualValues(t, 0, res[0].StmtStart)
		assert.EqualValues(t, 58, res[1].StmtStart)
	})

	t.Run("Should not return statements that are safe", func(t *testing.T) {
		t.Parallel()

		var b strings.Builder
		b.WriteString("DROP TABLE IF EXISTS pgvet;\n")
		b.WriteString("ALTER TABLE pgvet DROP COLUMN IF EXISTS value;\n")
		b.WriteString("DROP INDEX IF EXISTS pgvet_idx;\n")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 3)

		res, err := missingIfExists(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		assert.Empty(t, res)
	})

	t.Run("Should find no violations", func(t *testing.T) {
		t.Parallel()

		var b strings.Builder
		b.WriteString("ALTER TABLE pgvet RENAME COLUMN value TO new_value;\n")
		b.WriteString("ALTER TABLE pgvet RENAME COLUMN value TO new_value;\n")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 2)

		res, err := missingIfExists(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		assert.Empty(t, res)
	})
}
