package rules

import (
	pgquery "github.com/pganalyze/pg_query_go/v6"
)

var nullabilityRules = []Rule{
	{
		Code:     "add-non-null-column",
		Slug:     "Adding a non-nullable column without a default will fail if the table is populated",
		Help:     "Make the column nullable or add a default",
		Fn:       addNonNullColumn,
		Category: nullability,
	},
	{
		Code:     "set-non-null-column",
		Slug:     "Altering a column to be non-nullable might fail if the column contains null values",
		Help:     "Ensure that the column does not contain any null values",
		Fn:       alterColumnNotNullable,
		Category: nullability,
	},
}

func addNonNullColumn(
	tree *pgquery.ParseResult,
	code Code,
	slug,
	help string,
	implicitTransaction bool,
) ([]Result, error) {
	var results []Result
	for _, decl := range FilterStatements[*pgquery.Node_AlterTableStmt](tree.Stmts) {
		for _, cmd := range decl.Stmt.AlterTableStmt.GetCmds() {
			subtype := cmd.GetAlterTableCmd().GetSubtype()
			if subtype != pgquery.AlterTableType_AT_AddColumn {
				continue
			}

			column := cmd.GetAlterTableCmd().GetDef().GetColumnDef()
			if column == nil {
				continue
			}

			var hasDefault bool
			var r *Result

			for _, constraint := range column.GetConstraints() {
				if constraint.GetConstraint().GetContype() == pgquery.ConstrType_CONSTR_DEFAULT {
					hasDefault = true
				}

				if constraint.GetConstraint().GetContype() == pgquery.ConstrType_CONSTR_NOTNULL {
					r = &Result{
						Slug:      slug,
						Help:      help,
						Code:      code,
						StmtStart: decl.Start,
						StmtEnd:   decl.End,
					}
				}
			}
			if !hasDefault && r != nil {
				results = append(results, *r)
			}
		}
	}
	return results, nil
}

func alterColumnNotNullable(
	tree *pgquery.ParseResult,
	code Code,
	slug,
	help string,
	implicitTransaction bool,
) ([]Result, error) {
	var results []Result
	for _, stmt := range tree.Stmts {
		for _, cmd := range stmt.Stmt.GetAlterTableStmt().GetCmds() {
			subtype := cmd.GetAlterTableCmd().GetSubtype()
			if subtype == pgquery.AlterTableType_AT_SetNotNull {
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
