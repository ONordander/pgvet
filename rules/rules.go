package rules

import (
	pgquery "github.com/pganalyze/pg_query_go/v6"
)

const (
	breaking    = "breaking change"
	nullability = "nullability"
	idempotency = "idempotency"
	locking     = "locking"
)

type Rule struct {
	Code         Code
	Slug         string
	Help         string
	Fn           func(*pgquery.ParseResult, Code, string, string) ([]Result, error)
	Category     string
	Experimental bool
}

func AllRules() []Rule {
	var rules []Rule
	rules = append(rules, breakingRules...)
	rules = append(rules, nullabilityRules...)
	rules = append(rules, lockingRules...)
	rules = append(rules, idempotencyRules...)
	return rules
}

type Code string

type Result struct {
	Slug      string
	Help      string
	Code      Code
	StmtStart int32
	StmtEnd   int32
}
