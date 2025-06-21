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

		tree := mustParse(t, "ALTER TABLE pgvet DROP COLUMN value;")
		require.Len(t, tree.Stmts, 1)

		res, err := dropColumn(tree, testCode, testSlug, testHelp, true)
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
		b.WriteString("ALTER TABLE pgvet DROP COLUMN value;\n")
		b.WriteString("ALTER TABLE pgvet RENAME COLUMN value TO new_value;\n")
		b.WriteString("ALTER TABLE pgvet DROP COLUMN fk;\n")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 3)

		res, err := dropColumn(tree, testCode, testSlug, testHelp, true)
		require.NoError(t, err)
		require.Len(t, res, 2)

		assert.EqualValues(t, 0, res[0].StmtStart)
		assert.EqualValues(t, 88, res[1].StmtStart)
	})

	t.Run("Should find no violations", func(t *testing.T) {
		t.Parallel()

		var b strings.Builder
		b.WriteString("CREATE INDEX CONCURRENTLY on pgvet (id);\n")
		b.WriteString("ALTER TABLE pgvet RENAME COLUMN value TO value2;\n")
		b.WriteString("DROP INDEX pgvet_idx;")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 3)

		res, err := dropColumn(tree, testCode, testSlug, testHelp, true)
		require.NoError(t, err)
		assert.Empty(t, res)
	})
}

func TestDropTable(t *testing.T) {
	t.Parallel()

	t.Run("Should find violation", func(t *testing.T) {
		t.Parallel()

		tree := mustParse(t, "DROP TABLE pgvet;")
		require.Len(t, tree.Stmts, 1)

		res, err := dropTable(tree, testCode, testSlug, testHelp, true)
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
		b.WriteString("DROP TABLE pgvet;\n")
		b.WriteString("ALTER TABLE pgvet RENAME COLUMN value TO new_value;\n")
		b.WriteString("DROP TABLE pgvet_two;\n")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 3)

		res, err := dropTable(tree, testCode, testSlug, testHelp, true)
		require.NoError(t, err)
		require.Len(t, res, 2)

		assert.EqualValues(t, 0, res[0].StmtStart)
		assert.EqualValues(t, 69, res[1].StmtStart)
	})
}

func TestRenameColumn(t *testing.T) {
	t.Parallel()

	t.Run("Should find violation", func(t *testing.T) {
		t.Parallel()

		tree := mustParse(t, "ALTER TABLE pgvet RENAME COLUMN value TO value2;")
		require.Len(t, tree.Stmts, 1)

		res, err := renameColumn(tree, testCode, testSlug, testHelp, true)
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
		b.WriteString("CREATE INDEX CONCURRENTLY on pgvet (id);\n")
		b.WriteString("ALTER TABLE pgvet RENAME COLUMN value TO value2;\n")
		b.WriteString("ALTER TABLE pgvet RENAME COLUMN id TO id_new;")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 3)

		res, err := renameColumn(tree, testCode, testSlug, testHelp, true)
		require.NoError(t, err)
		require.Len(t, res, 2)

		assert.EqualValues(t, 40, res[0].StmtStart)
		assert.EqualValues(t, 89, res[1].StmtStart)
	})
}

func TestRenameTable(t *testing.T) {
	t.Parallel()

	t.Run("Should find violation", func(t *testing.T) {
		t.Parallel()

		tree := mustParse(t, "ALTER TABLE pgvet RENAME TO pgvet_new;")
		require.Len(t, tree.Stmts, 1)

		res, err := renameTable(tree, testCode, testSlug, testHelp, true)
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
		b.WriteString("CREATE INDEX CONCURRENTLY on pgvet (id);\n")
		b.WriteString("ALTER TABLE pgvet RENAME TO pgvet_new;\n")
		b.WriteString("ALTER TABLE otherpgvet RENAME TO otherpgvet_new;")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 3)

		res, err := renameTable(tree, testCode, testSlug, testHelp, true)
		require.NoError(t, err)
		require.Len(t, res, 2)

		assert.EqualValues(t, 40, res[0].StmtStart)
		assert.EqualValues(t, 79, res[1].StmtStart)
	})
}

func TestChangeColumnType(t *testing.T) {
	t.Parallel()

	t.Run("Should find violation", func(t *testing.T) {
		t.Parallel()

		tree := mustParse(t, "ALTER TABLE pgvet ALTER COLUMN value TYPE text;")
		require.Len(t, tree.Stmts, 1)

		res, err := changeColumnType(tree, testCode, testSlug, testHelp, true)
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
		b.WriteString("ALTER TABLE pgvet ALTER COLUMN value TYPE text;\n")
		b.WriteString("ALTER TABLE pgvet RENAME COLUMN value TO value2;\n")
		b.WriteString("ALTER TABLE pgvet ALTER COLUMN value TYPE varchar(36);\n")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 3)

		res, err := changeColumnType(tree, testCode, testSlug, testHelp, true)
		require.NoError(t, err)
		require.Len(t, res, 2)

		assert.EqualValues(t, 0, res[0].StmtStart)
		assert.EqualValues(t, 96, res[1].StmtStart)
	})
}
