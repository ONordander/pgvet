package rules

import (
	pgquery "github.com/pganalyze/pg_query_go/v6"
)

var lockingRules = []Rule{
	{
		Code:     "non-concurrent-index-creation",
		Slug:     "Creating an index non-concurrently acquires a lock on the table that block writes while the index is being built",
		Help:     "Build the index concurrently to avoid blocking. Note: this cannot be done inside a transaction",
		Fn:       nonConcurrentIndexCreation,
		Category: locking,
	},
}

func nonConcurrentIndexCreation(tree *pgquery.ParseResult, code Code, slug, help string) ([]Result, error) {
	var results []Result
	for _, decl := range FilterStatements[*pgquery.Node_IndexStmt](tree.Stmts) {
		if !decl.Stmt.IndexStmt.Concurrent {
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
