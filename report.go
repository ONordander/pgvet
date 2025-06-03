package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/onordander/pgvet/rules"
)

const (
	violationFmt = `%s%s%s: %s:%d

%s
  %sViolation%s: %s
  %sSolution%s: %s
  %sExplanation%s: https://github.com/ONordander/pgvet?tab=readme-ov-file#%s
%s
`
)

type violation struct {
	File          string     `json:"file"`
	Code          rules.Code `json:"code"`
	Statement     string     `json:"statement"`
	StatementLine int        `json:"statementLine"`
	Slug          string     `json:"slug"`
	Help          string     `json:"help"`
}

type Report []violation

func (r Report) Serialize(format string) (string, error) {
	if format == formatJson {
		var b strings.Builder
		if err := json.NewEncoder(&b).Encode(r); err != nil {
			return "", err
		}
		return b.String(), nil
	}

	var b strings.Builder
	files := map[string]bool{}
	for _, v := range r {
		b.WriteString(formatViolation(v))
		b.WriteString("\n")
		files[v.File] = true
	}

	summary := fmt.Sprintf("%s0 violations found %s\n", green, normal)
	if len(r) > 0 {
		summary = fmt.Sprintf("%s%d violation(s) found in %d file(s)%s\n", red, len(r), len(files), normal)
	}
	b.WriteString(summary)

	return b.String(), nil
}

func formatViolation(v violation) string {
	return fmt.Sprintf(
		violationFmt,
		red, v.Code, normal, v.File, v.StatementLine,
		formatStatement(v.Statement, v.StatementLine),
		bold, normal, v.Slug,
		bold, normal, v.Help,
		bold, normal, v.Code,
		strings.Repeat(".", 120),
	)
}

func formatStatement(stmt string, linestart int) string {
	lines := strings.Split(strings.ReplaceAll(stmt, "\r\n", "\n"), "\n")
	var msg strings.Builder
	for i, line := range lines {
		msg.WriteString(fmt.Sprintf("  %d | %s\n", linestart+i, line))
	}
	return msg.String()
}
