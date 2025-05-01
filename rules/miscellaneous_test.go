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

		tree := mustParse(t, "CREATE TABLE pgcheck (reference text REFERENCES parent(id));")
		require.Len(t, tree.Stmts, 1)

		res, err := missingForeignKeyIndex(tree, testCode, testSlug, testHelp)
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

		tree := mustParse(t, "ALTER TABLE pgcheck ADD CONSTRAINT reference_fk FOREIGN KEY (reference) REFERENCES parent(id);")
		require.Len(t, tree.Stmts, 1)

		res, err := missingForeignKeyIndex(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		require.Len(t, res, 1)
	})

	t.Run("Should find multiple violations", func(t *testing.T) {
		t.Parallel()

		var b strings.Builder
		b.WriteString("CREATE TABLE pgcheck (reference text REFERENCES parent(id));\n")
		b.WriteString("ALTER TABLE pgcheck RENAME COLUMN value TO new_value;\n")
		b.WriteString("ALTER TABLE pgcheck ADD CONSTRAINT other_reference_fk FOREIGN KEY (other_reference) REFERENCES parent(id);")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 3)

		res, err := missingForeignKeyIndex(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		require.Len(t, res, 2)

		assert.EqualValues(t, 0, res[0].StmtStart)
		assert.EqualValues(t, 114, res[1].StmtStart)
	})

	t.Run("Should not find references that have indexes", func(t *testing.T) {
		t.Parallel()

		var b strings.Builder
		b.WriteString("CREATE TABLE pgcheck (reference text REFERENCES parent(id));\n")
		b.WriteString("CREATE INDEX ON pgcheck(reference);\n")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 2)

		res, err := missingForeignKeyIndex(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		assert.Empty(t, res)
	})
}
