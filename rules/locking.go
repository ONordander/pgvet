package rules

import (
	pgquery "github.com/pganalyze/pg_query_go/v6"
)

var lockingRules = []Rule{
	{
		Code:     "non-concurrent-index",
		Slug:     "Creating/dropping an index non-concurrently acquires a lock on the table that block writes for the duration of the operation",
		Help:     "Create/drop the index concurrently using the `CONCURRENTLY` option to avoid blocking. Note: this cannot be done inside a transaction",
		Fn:       nonConcurrentIndex,
		Category: locking,
	},
	{
		Code:     "constraint-excessive-lock",
		Slug:     "Adding a constraint acquires a lock blocking any writes during the constraint validation",
		Help:     "Append the `NOT VALID` option and then in a following transaction perform `ALTER TABLE VALIDATE CONSTRAINT ...`",
		Fn:       constraintExcessiveLock,
		Category: locking,
	},
	{
		Code:              "multiple-locks",
		Slug:              "Experimental: acquiring multiple locks in a single transaction can cause a deadlock.",
		Help:              "Perform the changes in separate transactions",
		Fn:                multipleLocks,
		Category:          locking,
		DisabledByDefault: true,
	},
}

func nonConcurrentIndex(
	tree *pgquery.ParseResult,
	code Code,
	slug,
	help string,
	implicitMigration bool,
) ([]Result, error) {
	var results []Result
	for _, stmt := range tree.Stmts {
		// Check for index creation
		if indexStmt := stmt.GetStmt().GetIndexStmt(); indexStmt != nil {
			if !indexStmt.Concurrent {
				r := Result{
					Slug:      slug,
					Help:      help,
					Code:      code,
					StmtStart: stmt.GetStmtLocation(),
					StmtEnd:   stmt.GetStmtLocation() + stmt.GetStmtLen(),
				}
				results = append(results, r)
			}
		}
		// Check for index creation
		if dropStmt := stmt.GetStmt().GetDropStmt(); dropStmt != nil {
			if dropStmt.Concurrent {
				continue
			}
			if dropStmt.GetRemoveType() == pgquery.ObjectType_OBJECT_INDEX {
				r := Result{
					Slug:      slug,
					Help:      help,
					Code:      code,
					StmtStart: stmt.GetStmtLocation(),
					StmtEnd:   stmt.GetStmtLocation() + stmt.GetStmtLen(),
				}
				results = append(results, r)
			}
		}
	}

	return results, nil
}

func constraintExcessiveLock(
	tree *pgquery.ParseResult,
	code Code,
	slug,
	help string,
	implicitMigration bool,
) ([]Result, error) {
	var results []Result
	for _, decl := range FilterStatements[*pgquery.Node_AlterTableStmt](tree.Stmts) {
		for _, cmd := range decl.Stmt.AlterTableStmt.GetCmds() {
			alterTableCmd := cmd.GetAlterTableCmd()

			isAddConstraint := alterTableCmd.GetSubtype() == pgquery.AlterTableType_AT_AddConstraint
			isInitiallyValid := alterTableCmd.GetDef().GetConstraint().GetInitiallyValid() // maps to NOT VALID

			if isAddConstraint && isInitiallyValid {
				r := Result{
					Slug:      slug,
					Help:      help,
					Code:      code,
					StmtStart: decl.Start,
					StmtEnd:   decl.End,
				}
				results = append(results, r)
			}
		}
	}

	return results, nil
}

func multipleLocks(
	tree *pgquery.ParseResult,
	code Code,
	slug,
	help string,
	implicitMigration bool,
) ([]Result, error) {
	var results []Result

	tracker := newTXTracker(implicitMigration)

	for _, stmt := range tree.Stmts {
		// Check for alter table
		if alterTableStmt := stmt.GetStmt().GetAlterTableStmt(); alterTableStmt != nil {
			if tracker.add(alterTableStmt.GetRelation().GetRelname()) > 1 {
				r := Result{
					Slug:      slug,
					Help:      help,
					Code:      code,
					StmtStart: stmt.GetStmtLocation(),
					StmtEnd:   stmt.GetStmtLocation() + stmt.GetStmtLen(),
				}
				results = append(results, r)
			}
		}

		// Check for new tx or commit
		if txStmt := stmt.GetStmt().GetTransactionStmt(); txStmt != nil {
			switch txStmt.GetKind() {
			case pgquery.TransactionStmtKind_TRANS_STMT_COMMIT:
				tracker.commitTx()
			case pgquery.TransactionStmtKind_TRANS_STMT_BEGIN, pgquery.TransactionStmtKind_TRANS_STMT_START:
				tracker.beginTx()
			}
		}
	}

	return results, nil
}

type txTracker struct {
	tableChanges map[string]bool
	inTx         bool
}

func newTXTracker(startInTx bool) *txTracker {
	tracker := &txTracker{
		tableChanges: map[string]bool{},
		inTx:         startInTx, // Assume that we start in a transaction
	}
	return tracker
}

// Returns the number of tables in the current tx after the add
func (t *txTracker) add(tableName string) int {
	if !t.inTx {
		return 0
	}
	t.tableChanges[tableName] = true
	return len(t.tableChanges)
}

func (t *txTracker) beginTx() {
	// If we're already in a TX this is a no-op with a warning
	t.inTx = true
}

func (t *txTracker) commitTx() {
	// If we're not in a TX this is a no-op with a warning
	t.tableChanges = map[string]bool{}
	t.inTx = false
}
