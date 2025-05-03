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

func TestConstraintKeyExcessiveLock(t *testing.T) {
	t.Parallel()

	t.Run("Should find single violation", func(t *testing.T) {
		t.Parallel()

		tree := mustParse(t, "ALTER TABLE pgcheck ADD CONSTRAINT reference_fk FOREIGN KEY (reference) REFERENCES issues(id);")
		require.Len(t, tree.Stmts, 1)

		res, err := constraintExcessiveLock(tree, testCode, testSlug, testHelp)
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
		b.WriteString("ALTER TABLE pgcheck ADD CONSTRAINT reference_fk FOREIGN KEY (reference) REFERENCES issues(id);\n")
		b.WriteString("ALTER TABLE pgcheck ADD COLUMN value text;\n")
		b.WriteString("ALTER TABLE pgcheck ADD CONSTRAINT parent_fk FOREIGN KEY (parent) REFERENCES parent(id);\n")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 3)

		res, err := constraintExcessiveLock(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		require.Len(t, res, 2)

		assert.EqualValues(t, 0, res[0].StmtStart)
		assert.EqualValues(t, 137, res[1].StmtStart)
	})

	t.Run("Should find not constraint that has NOT VALID", func(t *testing.T) {
		t.Parallel()

		tree := mustParse(t, "ALTER TABLE pgcheck ADD CONSTRAINT reference_fk FOREIGN KEY (reference) REFERENCES issues(id) NOT VALID;")
		require.Len(t, tree.Stmts, 1)

		res, err := constraintExcessiveLock(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		assert.Empty(t, res)
	})
}

func TestManyAlterTable(t *testing.T) {
	t.Parallel()

	t.Run("Should find single violation", func(t *testing.T) {
		t.Parallel()

		var b strings.Builder
		b.WriteString("BEGIN;\n")
		b.WriteString("ALTER TABLE pgcheck ADD COLUMN value text;\n")
		b.WriteString("ALTER TABLE othertable ADD COLUMN value text;\n")
		b.WriteString("COMMIT;\n")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 4)

		res, err := manyAlterTable(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		require.Len(t, res, 1)

		assert.EqualValues(t, 49, res[0].StmtStart)
		assert.Greater(t, res[0].StmtEnd, res[0].StmtStart)
		assert.Equal(t, testCode, res[0].Code)
		assert.Equal(t, testSlug, res[0].Slug)
		assert.Equal(t, testHelp, res[0].Help)
	})

	t.Run("Should find multiple violations", func(t *testing.T) {
		t.Parallel()

		var b strings.Builder
		b.WriteString("BEGIN;\n")
		b.WriteString("ALTER TABLE pgcheck ADD COLUMN value text;\n")
		b.WriteString("ALTER TABLE othertable ADD COLUMN value text;\n")
		b.WriteString("ALTER TABLE thirdtable ADD COLUMN value text;\n")
		b.WriteString("COMMIT;\n")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 5)

		res, err := manyAlterTable(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		require.Len(t, res, 2)

		assert.EqualValues(t, 49, res[0].StmtStart)
		assert.EqualValues(t, 95, res[1].StmtStart)
	})

	t.Run("Should assume that the query starts in a transaction", func(t *testing.T) {
		t.Parallel()

		var b strings.Builder
		b.WriteString("ALTER TABLE pgcheck ADD COLUMN value text;\n")
		b.WriteString("ALTER TABLE othertable ADD COLUMN value text;\n")
		b.WriteString("COMMIT;\n")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 3)

		res, err := manyAlterTable(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("Should not find changes in separate transactions", func(t *testing.T) {
		t.Parallel()

		var b strings.Builder
		b.WriteString("BEGIN;\n")
		b.WriteString("ALTER TABLE pgcheck ADD COLUMN value text;\n")
		b.WriteString("COMMIT;\n")
		b.WriteString("BEGIN;\n")
		b.WriteString("ALTER TABLE othertable ADD COLUMN value text;\n")
		b.WriteString("COMMIT;\n")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 6)

		res, err := manyAlterTable(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		assert.Len(t, res, 0)
	})

	t.Run("Should not find changes outside of transactions", func(t *testing.T) {
		t.Parallel()

		var b strings.Builder
		b.WriteString("BEGIN;\n")
		b.WriteString("ALTER TABLE pgcheck ADD COLUMN value text;\n")
		b.WriteString("COMMIT;\n")
		b.WriteString("ALTER TABLE othertable ADD COLUMN value text;\n")
		b.WriteString("ALTER TABLE thirdtable ADD COLUMN value text;\n")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 5)

		res, err := manyAlterTable(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		assert.Len(t, res, 0)
	})

	t.Run("Should work with END/START TRANSACTION", func(t *testing.T) {
		t.Parallel()

		var b strings.Builder
		b.WriteString("BEGIN;\n")
		b.WriteString("ALTER TABLE pgcheck ADD COLUMN value text;\n")
		b.WriteString("END TRANSACTION;\n")
		b.WriteString("START TRANSACTION;\n")
		b.WriteString("ALTER TABLE othertable ADD COLUMN value text;\n")
		b.WriteString("COMMIT;\n")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 6)

		res, err := manyAlterTable(tree, testCode, testSlug, testHelp)
		require.NoError(t, err)
		assert.Len(t, res, 0)
	})
}
