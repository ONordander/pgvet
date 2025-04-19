package rules

import (
	pgquery "github.com/pganalyze/pg_query_go/v6"
)

type StmtDecl[T any] struct {
	Stmt  T
	Start int32
	End   int32
}

func FilterStatements[T any](rawStmts []*pgquery.RawStmt) []StmtDecl[T] {
	var filtered []StmtDecl[T]
	for _, raw := range rawStmts {
		stmt := raw.GetStmt()
		if stmt == nil {
			continue
		}
		if target, ok := stmt.GetNode().(T); ok {
			filtered = append(filtered, StmtDecl[T]{
				Stmt:  target,
				Start: raw.StmtLocation,
				End:   raw.StmtLocation + raw.StmtLen,
			})
		}
	}
	return filtered
}
