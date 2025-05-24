package rules

import (
	pgquery "github.com/pganalyze/pg_query_go/v6"
)

var idempotencyRules = []Rule{
	{
		Code:     "missing-if-not-exists",
		Slug:     "Creating/altering a relation might fail if it already exists, making the migration non idempotent",
		Help:     "Wrap the create statements with guards; e.g. CREATE TABLE IF NOT EXISTS pgvet ...",
		Fn:       missingIfNotExists,
		Category: idempotency,
	},
	{
		Code:     "missing-if-exists",
		Slug:     "Dropping an object/relation might fail if it doesn't exist, making the migration non idempotent",
		Help:     "Wrap the statements with guards; e.g. DROP INDEX CONCURRENTLY IF EXISTS pgvet_idx",
		Fn:       missingIfExists,
		Category: idempotency,
	},
}

func missingIfNotExists(tree *pgquery.ParseResult, code Code, slug, help string) ([]Result, error) {
	var results []Result
	for _, stmt := range tree.Stmts {
		// Check relation creations
		if createStmt := stmt.GetStmt().GetCreateStmt(); createStmt != nil {
			if !createStmt.IfNotExists {
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

		// Check add columns
		if alterTableStmt := stmt.GetStmt().GetAlterTableStmt(); alterTableStmt != nil {
			for _, cmd := range alterTableStmt.GetCmds() {
				isAddColumn := cmd.GetAlterTableCmd().GetSubtype() == pgquery.AlterTableType_AT_AddColumn
				missingIfNotExists := cmd.GetAlterTableCmd().GetMissingOk()
				if isAddColumn && !missingIfNotExists {
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

		// Check index creations
		if indexStmt := stmt.GetStmt().GetIndexStmt(); indexStmt != nil {
			isNamedIndex := indexStmt.Idxname != ""
			if !indexStmt.IfNotExists && isNamedIndex {
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

func missingIfExists(tree *pgquery.ParseResult, code Code, slug, help string) ([]Result, error) {
	var results []Result
	for _, stmt := range tree.Stmts {
		// Check drop relations, e.g. DROP TABLE, DROP INDEX
		if dropStmt := stmt.GetStmt().GetDropStmt(); dropStmt != nil {
			if !dropStmt.MissingOk {
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

		// Check drop columns
		if alterTableStmt := stmt.GetStmt().GetAlterTableStmt(); alterTableStmt != nil {
			for _, cmd := range alterTableStmt.GetCmds() {
				isDropColumn := cmd.GetAlterTableCmd().GetSubtype() == pgquery.AlterTableType_AT_DropColumn
				missingIfExists := cmd.GetAlterTableCmd().GetMissingOk()
				if isDropColumn && !missingIfExists {
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
	}

	return results, nil
}
