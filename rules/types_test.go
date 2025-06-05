package rules

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUseTimestampWithTimeZone(t *testing.T) {
	t.Parallel()

	t.Run("Should find create table violation", func(t *testing.T) {
		t.Parallel()

		tree := mustParse(t, "CREATE TABLE pgvet (created_at timestamp);")
		require.Len(t, tree.Stmts, 1)

		res, err := useTimestampWithTimeZone(tree, testCode, testSlug, testHelp, true)
		require.NoError(t, err)
		require.Len(t, res, 1)

		assert.EqualValues(t, 0, res[0].StmtStart)
		assert.EqualValues(t, 41, res[0].StmtEnd)
		assert.Equal(t, testCode, res[0].Code)
		assert.Equal(t, testSlug, res[0].Slug)
		assert.Equal(t, testHelp, res[0].Help)
	})

	t.Run("Should find add column violation", func(t *testing.T) {
		t.Parallel()

		tree := mustParse(t, "ALTER TABLE pgvet ADD COLUMN created_at timestamp;")
		require.Len(t, tree.Stmts, 1)

		res, err := useTimestampWithTimeZone(tree, testCode, testSlug, testHelp, true)
		require.NoError(t, err)
		require.Len(t, res, 1)

		assert.EqualValues(t, 0, res[0].StmtStart)
		assert.EqualValues(t, 49, res[0].StmtEnd)
		assert.Equal(t, testCode, res[0].Code)
		assert.Equal(t, testSlug, res[0].Slug)
		assert.Equal(t, testHelp, res[0].Help)
	})

	t.Run("Should find multiple violations", func(t *testing.T) {
		t.Parallel()

		var b strings.Builder
		b.WriteString("CREATE TABLE pgvet (created_at timestamp);\n")
		b.WriteString("ALTER TABLE pgvet RENAME COLUMN value TO new_value;\n")
		b.WriteString("ALTER TABLE pgvet ADD COLUMN updated_at timestamp;\n")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 3)

		res, err := useTimestampWithTimeZone(tree, testCode, testSlug, testHelp, true)
		require.NoError(t, err)
		require.Len(t, res, 2)

		assert.EqualValues(t, 0, res[0].StmtStart)
		assert.EqualValues(t, 94, res[1].StmtStart)
	})

	t.Run("Should find no violations for time zoned fields", func(t *testing.T) {
		t.Parallel()

		var b strings.Builder
		b.WriteString("CREATE TABLE pgvet (created_at timestamp with time zone);\n")
		b.WriteString("ALTER TABLE pgvet ADD COLUMN updated_at timestamptz;\n")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 2)

		res, err := dropColumn(tree, testCode, testSlug, testHelp, true)
		require.NoError(t, err)
		assert.Empty(t, res)
	})
}
