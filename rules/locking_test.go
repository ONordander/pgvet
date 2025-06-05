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

		tree := mustParse(t, "CREATE INDEX ON pgvet (id);")
		require.Len(t, tree.Stmts, 1)

		res, err := nonConcurrentIndex(tree, testCode, testSlug, testHelp, true)
		require.NoError(t, err)
		require.Len(t, res, 1)

		assert.EqualValues(t, 0, res[0].StmtStart)
		assert.Greater(t, res[0].StmtEnd, res[0].StmtStart)
		assert.Equal(t, testCode, res[0].Code)
		assert.Equal(t, testSlug, res[0].Slug)
		assert.Equal(t, testHelp, res[0].Help)
	})

	t.Run("Should find drop violation", func(t *testing.T) {
		t.Parallel()

		tree := mustParse(t, "DROP INDEX pgvet_idx;")
		require.Len(t, tree.Stmts, 1)

		res, err := nonConcurrentIndex(tree, testCode, testSlug, testHelp, true)
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
		b.WriteString("CREATE INDEX ON pgvet (id);\n")
		b.WriteString("ALTER TABLE pgvet ADD COLUMN value text;\n")
		b.WriteString("DROP INDEX pgvet_id_idx;\n")
		b.WriteString("DROP TABLE pgvet_prev;")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 4)

		res, err := nonConcurrentIndex(tree, testCode, testSlug, testHelp, true)
		require.NoError(t, err)
		require.Len(t, res, 2)

		assert.EqualValues(t, 0, res[0].StmtStart)
		assert.EqualValues(t, 68, res[1].StmtStart)
	})

	t.Run("Should find no violations", func(t *testing.T) {
		t.Parallel()

		var b strings.Builder
		b.WriteString("CREATE INDEX CONCURRENTLY on pgvet (id);\n")
		b.WriteString("ALTER TABLE pgvet ADD COLUMN value text;\n")
		b.WriteString("ALTER TABLE pgvet RENAME COLUMN value TO value2;\n")
		b.WriteString("DROP INDEX CONCURRENTLY pgvet_idx;")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 4)

		res, err := nonConcurrentIndex(tree, testCode, testSlug, testHelp, true)
		require.NoError(t, err)
		assert.Empty(t, res)
	})

	t.Run("Should not flag concurrent operations", func(t *testing.T) {
		t.Parallel()

		tree := mustParse(t, "CREATE INDEX CONCURRENTLY ON pgvet (id);")
		require.Len(t, tree.Stmts, 1)

		res, err := nonConcurrentIndex(tree, testCode, testSlug, testHelp, true)
		require.NoError(t, err)
		assert.Empty(t, res)
	})
}

func TestConstraintKeyExcessiveLock(t *testing.T) {
	t.Parallel()

	t.Run("Should find single violation", func(t *testing.T) {
		t.Parallel()

		tree := mustParse(t, "ALTER TABLE pgvet ADD CONSTRAINT reference_fk FOREIGN KEY (reference) REFERENCES issues(id);")
		require.Len(t, tree.Stmts, 1)

		res, err := constraintExcessiveLock(tree, testCode, testSlug, testHelp, true)
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
		b.WriteString("ALTER TABLE pgvet ADD CONSTRAINT reference_fk FOREIGN KEY (reference) REFERENCES issues(id);\n")
		b.WriteString("ALTER TABLE pgvet ADD COLUMN value text;\n")
		b.WriteString("ALTER TABLE pgvet ADD CONSTRAINT parent_fk FOREIGN KEY (parent) REFERENCES parent(id);\n")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 3)

		res, err := constraintExcessiveLock(tree, testCode, testSlug, testHelp, true)
		require.NoError(t, err)
		require.Len(t, res, 2)

		assert.EqualValues(t, 0, res[0].StmtStart)
		assert.EqualValues(t, 133, res[1].StmtStart)
	})

	t.Run("Should find not constraint that has NOT VALID", func(t *testing.T) {
		t.Parallel()

		tree := mustParse(t, "ALTER TABLE pgvet ADD CONSTRAINT reference_fk FOREIGN KEY (reference) REFERENCES issues(id) NOT VALID;")
		require.Len(t, tree.Stmts, 1)

		res, err := constraintExcessiveLock(tree, testCode, testSlug, testHelp, true)
		require.NoError(t, err)
		assert.Empty(t, res)
	})
}

func TestMultipleLocks(t *testing.T) {
	t.Parallel()

	t.Run("Should find single violation", func(t *testing.T) {
		t.Parallel()

		var b strings.Builder
		b.WriteString("BEGIN;\n")
		b.WriteString("ALTER TABLE pgvet ADD COLUMN value text;\n")
		b.WriteString("ALTER TABLE othertable ADD COLUMN value text;\n")
		b.WriteString("COMMIT;\n")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 4)

		res, err := multipleLocks(tree, testCode, testSlug, testHelp, true)
		require.NoError(t, err)
		require.Len(t, res, 1)

		assert.EqualValues(t, 47, res[0].StmtStart)
		assert.Greater(t, res[0].StmtEnd, res[0].StmtStart)
		assert.Equal(t, testCode, res[0].Code)
		assert.Equal(t, testSlug, res[0].Slug)
		assert.Equal(t, testHelp, res[0].Help)
	})

	t.Run("Should find multiple violations", func(t *testing.T) {
		t.Parallel()

		var b strings.Builder
		b.WriteString("BEGIN;\n")
		b.WriteString("ALTER TABLE pgvet ADD COLUMN value text;\n")
		b.WriteString("ALTER TABLE othertable ADD COLUMN value text;\n")
		b.WriteString("ALTER TABLE thirdtable ADD COLUMN value text;\n")
		b.WriteString("COMMIT;\n")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 5)

		res, err := multipleLocks(tree, testCode, testSlug, testHelp, true)
		require.NoError(t, err)
		require.Len(t, res, 2)

		assert.EqualValues(t, 47, res[0].StmtStart)
		assert.EqualValues(t, 93, res[1].StmtStart)
	})

	t.Run("Should assume that the query starts in a transaction", func(t *testing.T) {
		t.Parallel()

		var b strings.Builder
		b.WriteString("ALTER TABLE pgvet ADD COLUMN value text;\n")
		b.WriteString("ALTER TABLE othertable ADD COLUMN value text;\n")
		b.WriteString("COMMIT;\n")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 3)

		res, err := multipleLocks(tree, testCode, testSlug, testHelp, true)
		require.NoError(t, err)
		assert.Len(t, res, 1)
	})

	t.Run("Should not find changes in separate transactions", func(t *testing.T) {
		t.Parallel()

		var b strings.Builder
		b.WriteString("BEGIN;\n")
		b.WriteString("ALTER TABLE pgvet ADD COLUMN value text;\n")
		b.WriteString("COMMIT;\n")
		b.WriteString("BEGIN;\n")
		b.WriteString("ALTER TABLE othertable ADD COLUMN value text;\n")
		b.WriteString("COMMIT;\n")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 6)

		res, err := multipleLocks(tree, testCode, testSlug, testHelp, true)
		require.NoError(t, err)
		assert.Empty(t, res)
	})

	t.Run("Should not find changes outside of transactions", func(t *testing.T) {
		t.Parallel()

		var b strings.Builder
		b.WriteString("BEGIN;\n")
		b.WriteString("ALTER TABLE pgvet ADD COLUMN value text;\n")
		b.WriteString("COMMIT;\n")
		b.WriteString("ALTER TABLE othertable ADD COLUMN value text;\n")
		b.WriteString("ALTER TABLE thirdtable ADD COLUMN value text;\n")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 5)

		res, err := multipleLocks(tree, testCode, testSlug, testHelp, true)
		require.NoError(t, err)
		assert.Empty(t, res)
	})

	t.Run("Should work with END/START TRANSACTION", func(t *testing.T) {
		t.Parallel()

		var b strings.Builder
		b.WriteString("BEGIN;\n")
		b.WriteString("ALTER TABLE pgvet ADD COLUMN value text;\n")
		b.WriteString("END TRANSACTION;\n")
		b.WriteString("START TRANSACTION;\n")
		b.WriteString("ALTER TABLE othertable ADD COLUMN value text;\n")
		b.WriteString("COMMIT;\n")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 6)

		res, err := multipleLocks(tree, testCode, testSlug, testHelp, true)
		require.NoError(t, err)
		assert.Empty(t, res)
	})
}
