package main

import (
	"slices"
	"strings"

	"github.com/onordander/pgvet/rules"
)

const marker = "-- pgvet_nolint:"

func filterNoLints(rawQuery string, results []rules.Result) []rules.Result {
	var filtered []rules.Result
	for _, result := range results {
		raw := strings.TrimSpace(rawQuery[result.StmtStart:result.StmtEnd])
		var noLintRules []string
		for _, l := range strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n") {
			if !strings.HasPrefix(l, marker) {
				continue
			}
			trimmed := strings.TrimPrefix(l, marker)
			fields := strings.Fields(trimmed)
			if len(fields) == 0 {
				continue
			}

			noLintRules = append(noLintRules, strings.Split(fields[0], ",")...)
		}
		if !slices.Contains(noLintRules, string(result.Code)) {
			filtered = append(filtered, result)
		}
	}

	return filtered
}
