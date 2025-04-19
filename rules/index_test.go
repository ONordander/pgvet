package rules

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNonConcurrentIndex(t *testing.T) {
	t.Parallel()

	t.Run("Should find single violation", func(t *testing.T) {
		t.Parallel()

		tree := mustParse(t, "CREATE INDEX ON pgcheck (id);")
		require.Len(t, tree.Stmts, 1)

		res, err := nonConcurrentIndexCreation(tree, testCode, testSlug, testHelp)
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
		b.WriteString("CREATE INDEX on pgcheck (id);\n")
		b.WriteString("ALTER TABLE pgcheck ADD COLUMN value text;\n")
		b.WriteString("CREATE INDEX on pgcheck (value);\n")
		b.WriteString("DROP TABLE pgcheck_prev;")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 4)

		res, err := nonConcurrentIndexCreation(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		require.Len(t, res, 2)

		assert.EqualValues(t, 0, res[0].StmtStart)
		assert.EqualValues(t, 72, res[1].StmtStart)
	})

	t.Run("Should find no violations", func(t *testing.T) {
		t.Parallel()

		var b strings.Builder
		b.WriteString("CREATE INDEX CONCURRENTLY on pgcheck (id);\n")
		b.WriteString("ALTER TABLE pgcheck ADD COLUMN value text;\n")
		b.WriteString("ALTER TABLE pgcheck RENAME COLUMN value TO value2;\n")
		b.WriteString("DROP INDEX pgcheck_idx;")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 4)

		res, err := nonConcurrentIndexCreation(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		assert.Empty(t, res)
	})

	t.Run("Should not flag concurrent creations", func(t *testing.T) {
		t.Parallel()

		tree := mustParse(t, "CREATE INDEX CONCURRENTLY ON pgcheck (id);")
		require.Len(t, tree.Stmts, 1)

		res, err := nonConcurrentIndexCreation(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		assert.Empty(t, res)
	})
}
