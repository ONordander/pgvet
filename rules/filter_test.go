package rules

import (
	"strings"
	"testing"

	pgquery "github.com/pganalyze/pg_query_go/v6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilterStatements(t *testing.T) {
	t.Parallel()

	t.Run("Should find single statement", func(t *testing.T) {
		t.Parallel()

		tree := mustParse(t, "ALTER TABLE pgvet ADD COLUMN id INTEGER;")
		require.Len(t, tree.Stmts, 1)

		filtered := FilterStatements[*pgquery.Node_AlterTableStmt](tree.Stmts)
		require.Len(t, filtered, 1)
		assert.NotNil(t, filtered[0].Stmt.AlterTableStmt)
	})

	t.Run("Should find multiple statements", func(t *testing.T) {
		t.Parallel()

		var b strings.Builder
		b.WriteString("ALTER TABLE pgvet ADD COLUMN id INTEGER;")
		b.WriteString("ALTER TABLE pgvet DROP COLUMN id;")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 2)

		filtered := FilterStatements[*pgquery.Node_AlterTableStmt](tree.Stmts)
		require.Len(t, filtered, 2)
		assert.NotNil(t, filtered[0].Stmt.AlterTableStmt)
		assert.NotNil(t, filtered[1].Stmt.AlterTableStmt)
		assert.NotEqual(t, filtered[0], filtered[1])
	})

	t.Run("Should filter statements", func(t *testing.T) {
		t.Parallel()

		var b strings.Builder
		b.WriteString("CREATE INDEX ON pgvet (id);")
		b.WriteString("ALTER TABLE pgvet DROP COLUMN id;")
		b.WriteString("CREATE TABLE pgvet2 (id INTEGER);")
		tree := mustParse(t, b.String())
		require.Len(t, tree.Stmts, 3)

		filtered := FilterStatements[*pgquery.Node_CreateStmt](tree.Stmts)
		require.Len(t, filtered, 1)
		assert.NotNil(t, filtered[0].Stmt.CreateStmt)
	})
}

func mustParse(t *testing.T, q string) *pgquery.ParseResult {
	t.Helper()

	tree, err := pgquery.Parse(q)
	require.NoError(t, err)

	return tree
}
