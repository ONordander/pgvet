package rules

import (
	"maps"
	"slices"

	pgquery "github.com/pganalyze/pg_query_go/v6"
)

var miscellaneousRules = []Rule{
	{
		Code:     "missing-foreign-key-index",
		Slug:     "PostgreSQL does not create an automatic index for foreign key constraints.",
		Help:     "Add an index for the foreign key constraint column",
		Fn:       missingForeignKeyIndex,
		Category: miscellaneous,
	},
	{
		Code:     "concurrent-in-tx",
		Slug:     "Concurrently creating/dropping an index cannot be done inside of a transaction",
		Help:     "Perform the operation outside of a transaction",
		Fn:       concurrentInTX,
		Category: miscellaneous,
	},
}

func missingForeignKeyIndex(
	tree *pgquery.ParseResult,
	code Code,
	slug,
	help string,
	implicitTransaction bool,
) ([]Result, error) {
	type stmtMarker struct {
		stmtStart int32
		stmtEnd   int32
	}

	unindexedConstraints := map[string]stmtMarker{}

	for _, stmt := range tree.Stmts {
		// Create table statement, check for FK constraints
		createStmt := stmt.GetStmt().GetCreateStmt()
		for _, col := range createStmt.GetTableElts() {
			for _, constraint := range col.GetColumnDef().GetConstraints() {
				if constraint.GetConstraint().GetContype() == pgquery.ConstrType_CONSTR_FOREIGN {
					tableName := createStmt.GetRelation().GetRelname()
					columnName := col.GetColumnDef().GetColname()
					entry := tableName + "." + columnName
					unindexedConstraints[entry] = stmtMarker{
						stmtStart: stmt.GetStmtLocation(),
						stmtEnd:   stmt.GetStmtLocation() + stmt.GetStmtLen(),
					}
				}
			}
		}

		// Alter table add constraint statement, check for FK constraints
		alterTableStmt := stmt.GetStmt().GetAlterTableStmt()
		for _, cmd := range alterTableStmt.GetCmds() {
			alterTableCmd := cmd.GetAlterTableCmd()

			isAddConstraint := alterTableCmd.GetSubtype() == pgquery.AlterTableType_AT_AddConstraint
			constraint := alterTableCmd.GetDef().GetConstraint()
			isForeignKey := constraint.GetContype() == pgquery.ConstrType_CONSTR_FOREIGN
			if isAddConstraint && isForeignKey {
				tableName := alterTableStmt.GetRelation().GetRelname()
				columnName := constraint.GetFkAttrs()[0].GetString_().GetSval()
				entry := tableName + "." + columnName
				unindexedConstraints[entry] = stmtMarker{
					stmtStart: stmt.GetStmtLocation(),
					stmtEnd:   stmt.GetStmtLocation() + stmt.GetStmtLen(),
				}
			}
		}

		// Create index statements, pop if index is found for FK
		indexStmt := stmt.GetStmt().GetIndexStmt()
		tableName := indexStmt.GetRelation().GetRelname()
		for _, param := range indexStmt.GetIndexParams() {
			entry := tableName + "." + param.GetIndexElem().GetName()
			delete(unindexedConstraints, entry)
		}
	}

	var results []Result
	sortedConstraints := slices.SortedFunc(maps.Values(unindexedConstraints), func(a, b stmtMarker) int {
		if a.stmtStart < b.stmtStart {
			return -1
		}
		return 1
	})

	for _, marker := range sortedConstraints {
		r := Result{
			Slug:      slug,
			Help:      help,
			Code:      code,
			StmtStart: marker.stmtStart,
			StmtEnd:   marker.stmtEnd,
		}
		results = append(results, r)
	}

	return results, nil
}

func concurrentInTX(
	tree *pgquery.ParseResult,
	code Code,
	slug,
	help string,
	implicitTransaction bool,
) ([]Result, error) {
	var results []Result
	tracker := newTXTracker(implicitTransaction)

	for _, stmt := range tree.Stmts {
		// Check for new tx or commit
		if txStmt := stmt.GetStmt().GetTransactionStmt(); txStmt != nil {
			switch txStmt.GetKind() {
			case pgquery.TransactionStmtKind_TRANS_STMT_COMMIT:
				tracker.commitTx()
			case pgquery.TransactionStmtKind_TRANS_STMT_BEGIN, pgquery.TransactionStmtKind_TRANS_STMT_START:
				tracker.beginTx()
			}
		}

		if !tracker.inTx {
			continue
		}

		// Check for index creation
		if indexStmt := stmt.GetStmt().GetIndexStmt(); indexStmt != nil {
			if !indexStmt.Concurrent {
				continue
			}

			r := Result{
				Slug:      slug,
				Help:      help,
				Code:      code,
				StmtStart: stmt.GetStmtLocation(),
				StmtEnd:   stmt.GetStmtLocation() + stmt.GetStmtLen(),
			}
			results = append(results, r)
		}

		// Check for index drop
		if dropStmt := stmt.GetStmt().GetDropStmt(); dropStmt != nil {
			if !dropStmt.Concurrent {
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
