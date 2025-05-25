package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/onordander/pgvet/rules"
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
	for _, entry := range r {
		b.WriteString(entry.formatMsg())
		b.WriteString("\n")
	}
	color := red
	if len(r) == 0 {
		color = green
	}
	b.WriteString(fmt.Sprintf("%s%d violations found%s\n", color, len(r), normal))

	return b.String(), nil
}

func (l violation) formatMsg() string {
	msg := `%s%s%s: %s:%d

%s
  %sViolation%s: %s
  %sSolution%s: %s
  %sExplanation%s: https://github.com/ONordander/pgvet?tab=readme-ov-file#%s
%s
`
	return fmt.Sprintf(
		msg,
		red, l.Code, normal, l.File, l.StatementLine,
		formatStatement(l.Statement, l.StatementLine),
		bold, normal, l.Slug,
		bold, normal, l.Help,
		bold, normal, l.Code,
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
