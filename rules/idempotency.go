package rules

import (
	pgquery "github.com/pganalyze/pg_query_go/v6"
)

var idempotencyRules = []Rule{
	{
		Code:     "missing-if-not-exists",
		Slug:     "Creating an object might fail if it already exists, making the migration non idempotent",
		Help:     "Wrap the create statements with guards; e.g. CREATE TABLE IF NOT EXISTS pgvet ...",
		Fn:       missingIfNotExists,
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
