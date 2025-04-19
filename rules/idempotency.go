package rules

import (
	pgquery "github.com/pganalyze/pg_query_go/v6"
)

var idempotencyRules = []Rule{
	{
		Code:     "missing-relation-if-not-exists",
		Slug:     "Creating a relation will fail if it already exists, making the migration non idempotent",
		Help:     "Wrap the create statements with guards; e.g. CREATE TABLE IF NOT EXISTS pgcheck ...",
		Fn:       missingRelationIfNotExists,
		Category: idempotency,
	},
	{
		Code:     "missing-index-if-not-exists",
		Slug:     "Creating a named index will fail if it already exists, making the migration non idempotent",
		Help:     "Wrap the create statements with guards; e.g. CREATE INDEX CONCURRENTLY IF NOT EXISTS pgcheck_idx ...",
		Fn:       missingIndexIfNotExists,
		Category: idempotency,
	},
}

func missingRelationIfNotExists(tree *pgquery.ParseResult, code Code, slug, help string) ([]Result, error) {
	var results []Result
	for _, decl := range FilterStatements[*pgquery.Node_CreateStmt](tree.Stmts) {
		if !decl.Stmt.CreateStmt.IfNotExists {
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

	return results, nil
}

func missingIndexIfNotExists(tree *pgquery.ParseResult, code Code, slug, help string) ([]Result, error) {
	var results []Result
	for _, decl := range FilterStatements[*pgquery.Node_IndexStmt](tree.Stmts) {
		isNamedIndex := decl.Stmt.IndexStmt.Idxname != ""
		if !decl.Stmt.IndexStmt.IfNotExists && isNamedIndex {
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

	return results, nil
}
