package rules

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMissingForeignKeyIndex(t *testing.T) {
	t.Parallel()

	t.Run("Should find violation from CREATE TABLE", func(t *testing.T) {
		t.Parallel()

		tree := mustParse(t, "CREATE TABLE pgvet (reference text REFERENCES parent(id));")
		require.Len(t, tree.Stmts, 1)

		res, err := missingForeignKeyIndex(tree, testCode, testSlug, testHelp, true)
		require.NoError(t, err)
		require.Len(t, res, 1)

		assert.EqualValues(t, 0, res[0].StmtStart)
		assert.Greater(t, res[0].StmtEnd, res[0].StmtStart)
		assert.Equal(t, testCode, res[0].Code)
		assert.Equal(t, testSlug, res[0].Slug)
		assert.Equal(t, testHelp, res[0].Help)
	})

	t.Run("Should find violation from ALTER TABLE ADD CONSTRAINT", func(t *testing.T) {
		t.Parallel()

		tree := mustParse(t, "ALTER TABLE pgvet ADD CONSTRAINT reference_fk FOREIGN KEY (reference) REFERENCES parent(id);")
		require.Len(t, tree.Stmts, 1)

		res, err := missingForeignKeyIndex(tree, testCode, testSlug, testHelp, true)
		require.NoError(t, err)
		require.Len(t, res, 1)
	})

	t.Run("Should find multiple violations", func(t *testing.T) {
		t.Parallel()

		var b strings.Builder
		b.WriteString("CREATE TABLE pgvet (reference text REFERENCES parent(id));\n")
		b.WriteString("ALTER TABLE pgvet RENAME COLUMN value TO new_value;\n")
		b.WriteString("ALTER TABLE pgvet ADD CONSTRAINT other_reference_fk FOREIGN KEY (other_reference) REFERENCES parent(id);")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 3)

		res, err := missingForeignKeyIndex(tree, testCode, testSlug, testHelp, true)
		require.NoError(t, err)
		require.Len(t, res, 2)

		assert.EqualValues(t, 0, res[0].StmtStart)
		assert.EqualValues(t, 110, res[1].StmtStart)
	})

	t.Run("Should not find references that have indexes", func(t *testing.T) {
		t.Parallel()

		var b strings.Builder
		b.WriteString("CREATE TABLE pgvet (reference text REFERENCES parent(id));\n")
		b.WriteString("CREATE INDEX ON pgvet(reference);\n")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 2)

		res, err := missingForeignKeyIndex(tree, testCode, testSlug, testHelp, true)
		require.NoError(t, err)
		assert.Empty(t, res)
	})
}

func TestConcurrentInTX(t *testing.T) {
	t.Parallel()

	t.Run("Should find violation from CREATE INDEX", func(t *testing.T) {
		t.Parallel()

		tree := mustParse(t, "CREATE INDEX CONCURRENTLY pgvet_idx ON pgvet(value);\n")
		require.Len(t, tree.Stmts, 1)

		res, err := concurrentInTX(tree, testCode, testSlug, testHelp, true)
		require.NoError(t, err)
		require.Len(t, res, 1)

		assert.EqualValues(t, 0, res[0].StmtStart)
		assert.Greater(t, res[0].StmtEnd, res[0].StmtStart)
		assert.Equal(t, testCode, res[0].Code)
		assert.Equal(t, testSlug, res[0].Slug)
		assert.Equal(t, testHelp, res[0].Help)
	})

	t.Run("Should find violation from DROP INDEX", func(t *testing.T) {
		t.Parallel()

		tree := mustParse(t, "DROP INDEX CONCURRENTLY pgvet_idx;\n")
		require.Len(t, tree.Stmts, 1)

		res, err := concurrentInTX(tree, testCode, testSlug, testHelp, true)
		require.NoError(t, err)
		require.Len(t, res, 1)

		assert.EqualValues(t, 0, res[0].StmtStart)
		assert.Greater(t, res[0].StmtEnd, res[0].StmtStart)
		assert.Equal(t, testCode, res[0].Code)
		assert.Equal(t, testSlug, res[0].Slug)
		assert.Equal(t, testHelp, res[0].Help)
	})

	t.Run("Should find multiple violations", func(t *testing.T) {
		t.Parallel()

		var b strings.Builder
		b.WriteString("DROP INDEX CONCURRENTLY pgvet_idx;\n")
		b.WriteString("CREATE INDEX CONCURRENTLY pgvet_idx ON pgvet(value);\n")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 2)

		res, err := concurrentInTX(tree, testCode, testSlug, testHelp, true)
		require.NoError(t, err)
		require.Len(t, res, 2)

		assert.EqualValues(t, 0, res[0].StmtStart)
		assert.EqualValues(t, 34, res[1].StmtStart)
	})

	t.Run("Should not flag non-concurrent", func(t *testing.T) {
		t.Parallel()

		tree := mustParse(t, "CREATE INDEX pgvet_idx ON pgvet(value);\n")
		require.Len(t, tree.Stmts, 1)

		res, err := concurrentInTX(tree, testCode, testSlug, testHelp, true)
		require.NoError(t, err)
		assert.Empty(t, res)
	})

	t.Run("Should not flag when not in TX", func(t *testing.T) {
		t.Parallel()

		var b strings.Builder
		b.WriteString("CREATE INDEX CONCURRENTLY ON pgvet(reference);\n")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 1)

		res, err := concurrentInTX(tree, testCode, testSlug, testHelp, false)
		require.NoError(t, err)
		assert.Empty(t, res)
	})
}
