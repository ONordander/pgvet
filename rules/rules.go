package rules

import (
	pgquery "github.com/pganalyze/pg_query_go/v6"
)

const (
	breaking      = "breaking change"
	nullability   = "nullability"
	idempotency   = "idempotency"
	locking       = "locking"
	miscellaneous = "miscellaneous"
	types         = "types"
)

type Rule struct {
	Code              Code
	Slug              string
	Help              string
	Fn                func(*pgquery.ParseResult, Code, string, string, bool) ([]Result, error)
	Category          string
	DisabledByDefault bool
}

func AllRules() []Rule {
	var rules []Rule
	rules = append(rules, breakingRules...)
	rules = append(rules, nullabilityRules...)
	rules = append(rules, lockingRules...)
	rules = append(rules, idempotencyRules...)
	rules = append(rules, miscellaneousRules...)
	rules = append(rules, typeRules...)
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
