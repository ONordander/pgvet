package rules

import (
	pgquery "github.com/pganalyze/pg_query_go/v6"
)

var typeRules = []Rule{
	{
		Code:     "use-timestamp-with-time-zone",
		Slug:     "Timestamp with time zone preserves the time zone information and makes the data easier to reason about",
		Help:     "Update fields to use `timestamptz`/`timestamp with time zone` instead of `timestamp`/`timestamp without time zone`",
		Fn:       useTimestampWithTimeZone,
		Category: "types",
	},
}

func useTimestampWithTimeZone(tree *pgquery.ParseResult, code Code, slug, help string) ([]Result, error) {
	var results []Result
	for _, stmt := range tree.Stmts {
		// Check for table creation
		createStmt := stmt.GetStmt().GetCreateStmt()
		for _, col := range createStmt.GetTableElts() {
			for _, nameNode := range col.GetColumnDef().GetTypeName().GetNames() {
				if name := nameNode.GetString_(); name != nil {
					if name.Sval == "timestamp" {
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

		// Check for add column
		alterTableStmt := stmt.GetStmt().GetAlterTableStmt()
		for _, cmd := range alterTableStmt.GetCmds() {
			if cmd.GetAlterTableCmd().GetSubtype() != pgquery.AlterTableType_AT_AddColumn {
				continue
			}
			for _, nameNode := range cmd.GetAlterTableCmd().GetDef().GetColumnDef().GetTypeName().GetNames() {
				if name := nameNode.GetString_(); name != nil {
					if name.Sval == "timestamp" {
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
	}
	return results, nil
}
