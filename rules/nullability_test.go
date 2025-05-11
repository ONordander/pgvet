package rules

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddNonNullColumn(t *testing.T) {
	t.Parallel()

	t.Run("Should find violation", func(t *testing.T) {
		t.Parallel()

		tree := mustParse(t, "ALTER TABLE pgvet ADD COLUMN value text NOT NULL;")
		require.Len(t, tree.Stmts, 1)

		res, err := addNonNullColumn(tree, testCode, testSlug, testHelp)
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
		b.WriteString("ALTER TABLE pgvet ADD COLUMN value text NOT NULL;\n")
		b.WriteString("ALTER TABLE pgvet RENAME COLUMN value TO new_value;\n")
		b.WriteString("ALTER TABLE pgvet ADD COLUMN fk integer NOT NULL;\n")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 3)

		res, err := addNonNullColumn(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		require.Len(t, res, 2)

		assert.EqualValues(t, 0, res[0].StmtStart)
		assert.EqualValues(t, 101, res[1].StmtStart)
	})

	t.Run("Should not flag if column has a default", func(t *testing.T) {
		t.Parallel()

		tree := mustParse(t, "ALTER TABLE pgvet ADD COLUMN value text NOT NULL DEFAULT '1';")
		require.Len(t, tree.Stmts, 1)

		res, err := addNonNullColumn(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		assert.Empty(t, res)
	})

	t.Run("Should not flag if column is nullable", func(t *testing.T) {
		t.Parallel()

		tree := mustParse(t, "ALTER TABLE pgvet ADD COLUMN value text;")
		require.Len(t, tree.Stmts, 1)

		res, err := addNonNullColumn(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		assert.Empty(t, res)
	})

	t.Run("Should find no violations", func(t *testing.T) {
		t.Parallel()

		var b strings.Builder
		b.WriteString("CREATE INDEX CONCURRENTLY on pgvet (id);\n")
		b.WriteString("ALTER TABLE pgvet RENAME COLUMN value TO value2;\n")
		b.WriteString("ALTER TABLE pgvet ALTER COLUMN value SET NOT NULL;\n")
		b.WriteString("DROP INDEX pgvet_idx;")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 4)

		res, err := addNonNullColumn(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		assert.Empty(t, res)
	})
}

func TestAlterColumnNotNullable(t *testing.T) {
	t.Parallel()

	t.Run("Should find violation", func(t *testing.T) {
		t.Parallel()

		tree := mustParse(t, "ALTER TABLE pgvet ALTER COLUMN value SET NOT NULL;")
		require.Len(t, tree.Stmts, 1)

		res, err := alterColumnNotNullable(tree, testCode, testSlug, testHelp)
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
		b.WriteString("ALTER TABLE pgvet ALTER COLUMN value SET NOT NULL;\n")
		b.WriteString("ALTER TABLE pgvet RENAME COLUMN value TO new_value;\n")
		b.WriteString("ALTER TABLE pgvet ALTER COLUMN fk SET NOT NULL;\n")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 3)

		res, err := alterColumnNotNullable(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		require.Len(t, res, 2)

		assert.EqualValues(t, 0, res[0].StmtStart)
		assert.EqualValues(t, 102, res[1].StmtStart)
	})

	t.Run("Should not flag if column is made nullable", func(t *testing.T) {
		t.Parallel()

		tree := mustParse(t, "ALTER TABLE pgvet ALTER COLUMN value DROP NOT NULL;")
		require.Len(t, tree.Stmts, 1)

		res, err := alterColumnNotNullable(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		assert.Empty(t, res)
	})

	t.Run("Should find no violations", func(t *testing.T) {
		t.Parallel()

		var b strings.Builder
		b.WriteString("CREATE INDEX CONCURRENTLY on pgvet (id);\n")
		b.WriteString("ALTER TABLE pgvet RENAME COLUMN value TO value2;\n")
		b.WriteString("DROP INDEX pgvet_idx;")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 3)

		res, err := alterColumnNotNullable(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		assert.Empty(t, res)
	})
}
