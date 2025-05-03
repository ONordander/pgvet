package rules

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testCode = Code("test")
	testSlug = "testslug"
	testHelp = "testHelp"
)

func TestDropColumn(t *testing.T) {
	t.Parallel()

	t.Run("Should find violation", func(t *testing.T) {
		t.Parallel()

		tree := mustParse(t, "ALTER TABLE pgcheck DROP COLUMN value;")
		require.Len(t, tree.Stmts, 1)

		res, err := dropColumn(tree, testCode, testSlug, testHelp)
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
		b.WriteString("ALTER TABLE pgcheck DROP COLUMN value;\n")
		b.WriteString("ALTER TABLE pgcheck RENAME COLUMN value TO new_value;\n")
		b.WriteString("ALTER TABLE pgcheck DROP COLUMN fk;\n")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 3)

		res, err := dropColumn(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		require.Len(t, res, 2)

		assert.EqualValues(t, 0, res[0].StmtStart)
		assert.EqualValues(t, 92, res[1].StmtStart)
	})

	t.Run("Should find no violations", func(t *testing.T) {
		t.Parallel()

		var b strings.Builder
		b.WriteString("CREATE INDEX CONCURRENTLY on pgcheck (id);\n")
		b.WriteString("ALTER TABLE pgcheck RENAME COLUMN value TO value2;\n")
		b.WriteString("DROP INDEX pgcheck_idx;")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 3)

		res, err := dropColumn(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		assert.Empty(t, res)
	})
}

func TestDropTable(t *testing.T) {
	t.Parallel()

	t.Run("Should find violation", func(t *testing.T) {
		t.Parallel()

		tree := mustParse(t, "DROP TABLE pgcheck;")
		require.Len(t, tree.Stmts, 1)

		res, err := dropTable(tree, testCode, testSlug, testHelp)
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
		b.WriteString("DROP TABLE pgcheck;\n")
		b.WriteString("ALTER TABLE pgcheck RENAME COLUMN value TO new_value;\n")
		b.WriteString("DROP TABLE pgcheck_two;\n")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 3)

		res, err := dropTable(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		require.Len(t, res, 2)

		assert.EqualValues(t, 0, res[0].StmtStart)
		assert.EqualValues(t, 73, res[1].StmtStart)
	})
}

func TestRenameColumn(t *testing.T) {
	t.Parallel()

	t.Run("Should find violation", func(t *testing.T) {
		t.Parallel()

		tree := mustParse(t, "ALTER TABLE pgcheck RENAME COLUMN value TO value2;")
		require.Len(t, tree.Stmts, 1)

		res, err := renameColumn(tree, testCode, testSlug, testHelp)
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
		b.WriteString("CREATE INDEX CONCURRENTLY on pgcheck (id);\n")
		b.WriteString("ALTER TABLE pgcheck RENAME COLUMN value TO value2;\n")
		b.WriteString("ALTER TABLE pgcheck RENAME COLUMN id TO id_new;")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 3)

		res, err := renameColumn(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		require.Len(t, res, 2)

		assert.EqualValues(t, 42, res[0].StmtStart)
		assert.EqualValues(t, 93, res[1].StmtStart)
	})
}

func TestChangeColumnType(t *testing.T) {
	t.Parallel()

	t.Run("Should find violation", func(t *testing.T) {
		t.Parallel()

		tree := mustParse(t, "ALTER TABLE pgcheck ALTER COLUMN value TYPE text;")
		require.Len(t, tree.Stmts, 1)

		res, err := changeColumnType(tree, testCode, testSlug, testHelp)
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
		b.WriteString("ALTER TABLE pgcheck ALTER COLUMN value TYPE text;\n")
		b.WriteString("ALTER TABLE pgcheck RENAME COLUMN value TO value2;\n")
		b.WriteString("ALTER TABLE pgcheck ALTER COLUMN value TYPE varchar(36);\n")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 3)

		res, err := changeColumnType(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		require.Len(t, res, 2)

		assert.EqualValues(t, 0, res[0].StmtStart)
		assert.EqualValues(t, 100, res[1].StmtStart)
	})
}
