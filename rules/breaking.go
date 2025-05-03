package rules

import (
	pgquery "github.com/pganalyze/pg_query_go/v6"
)

var breakingRules = []Rule{
	{
		Code:     "drop-column",
		Slug:     "Dropping a column is not backwards compatible and may break existing clients",
		Help:     "Update the application code to no longer use the column before applying the change",
		Fn:       dropColumn,
		Category: breaking,
	},
	{
		Code:     "drop-table",
		Slug:     "Dropping a table is not backwards compatible and may break existing clients",
		Help:     "Update the application code to no longer use the table before applying the change",
		Fn:       dropTable,
		Category: breaking,
	},
	{
		Code:     "rename-column",
		Slug:     "Renaming a column is not backwards compatible and may break existing clients",
		Help:     "Add the new column as nullable and write to both from the application. Perform a backfill. Update application code to only use the new column. Delete the old column",
		Fn:       renameColumn,
		Category: breaking,
	},
	{
		Code:     "change-column-type",
		Slug:     "Changing the type of a column is not backwards compatible and may break existing clients",
		Help:     "Add a new column with the new type and write to both from the application. Perform a backfill. Update application code to only use the new column. Delete the old column",
		Fn:       changeColumnType,
		Category: breaking,
	},
}

func dropColumn(tree *pgquery.ParseResult, code Code, slug, help string) ([]Result, error) {
	var results []Result
	for _, decl := range FilterStatements[*pgquery.Node_AlterTableStmt](tree.Stmts) {
		for _, cmd := range decl.Stmt.AlterTableStmt.GetCmds() {
			subtype := cmd.GetAlterTableCmd().GetSubtype()
			if subtype != pgquery.AlterTableType_AT_DropColumn {
				continue
			}
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

func renameColumn(tree *pgquery.ParseResult, code Code, slug, help string) ([]Result, error) {
	var results []Result
	for _, decl := range FilterStatements[*pgquery.Node_RenameStmt](tree.Stmts) {
		if decl.Stmt.RenameStmt.RenameType != pgquery.ObjectType_OBJECT_COLUMN {
			continue
		}
		r := Result{
			Slug:      slug,
			Help:      help,
			Code:      code,
			StmtStart: decl.Start,
			StmtEnd:   decl.End,
		}
		results = append(results, r)
	}
	return results, nil
}

func dropTable(tree *pgquery.ParseResult, code Code, slug, help string) ([]Result, error) {
	var results []Result
	for _, decl := range FilterStatements[*pgquery.Node_DropStmt](tree.Stmts) {
		if decl.Stmt.DropStmt.RemoveType == pgquery.ObjectType_OBJECT_TABLE {
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

func changeColumnType(tree *pgquery.ParseResult, code Code, slug, help string) ([]Result, error) {
	var results []Result
	for _, decl := range FilterStatements[*pgquery.Node_AlterTableStmt](tree.Stmts) {
		for _, cmd := range decl.Stmt.AlterTableStmt.GetCmds() {
			alterTableCmd := cmd.GetAlterTableCmd()
			if alterTableCmd.GetSubtype() == pgquery.AlterTableType_AT_AlterColumnType {
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
