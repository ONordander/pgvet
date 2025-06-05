package main

import (
	"cmp"
	_ "embed"
	"flag"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/onordander/pgvet/rules"

	pgquery "github.com/wasilibs/go-pgquery"
)

//go:embed NOTICE.txt
var notice string

//go:embed VERSION.txt
var version string

const (
	formatJson = "json"
	formatText = "text"
)

func main() {
	wOut, wErr := os.Stdout, os.Stderr

	flagSet := flag.NewFlagSet("lint", flag.ExitOnError)
	flagSet.SetOutput(wErr)
	format := flagSet.String("format", "text", "Set output format, text or json. Text is default")
	exitStatusOnViolations := flagSet.Bool("exit-status-on-violation", false, "Set exit status >0 if any violations are found")
	config := flagSet.String("config", "", "Config file")
	flagSet.Usage = func() {
		fmt.Fprint(wErr, "Usage:\n")
		fmt.Fprint(wErr, "\t./pgvet lint [--config <config.yaml>] <filepattern>...\n")
		fmt.Fprint(wErr, "\t./pgvet --help\n")
		fmt.Fprint(wErr, "\t./pgvet version\n")
		fmt.Fprint(wErr, "\t./pgvet license\n")
		flagSet.PrintDefaults()
		fmt.Fprint(wErr, "Example:\n")
		fmt.Fprint(wErr, "\t./pgvet lint --config=config.yaml migrations/*.sql\n")
	}

	if len(os.Args) < 2 {
		flagSet.Usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "license":
		fmt.Fprint(wOut, notice)
		return
	case "version":
		fmt.Fprintf(wOut, "pgvet %s\n", strings.TrimSpace(version))
		return
	case "rules":
		for _, rule := range rules.AllRules() {
			enabled := "✓"
			if rule.DisabledByDefault {
				enabled = "🗙"
			}
			fmt.Fprintf(wOut, "%s%s%s\n", magenta, rule.Code, normal)
			fmt.Fprintf(wOut, "\t| %s%s%s\n", bold, rule.Slug, normal)
			fmt.Fprintf(wOut, "\tHelp: %s\n", rule.Help)
			fmt.Fprintf(wOut, "\tEnabled by default: %s\n", enabled)
			fmt.Fprintf(wOut, "\tExplanation: https://github.com/ONordander/pgvet?tab=readme-ov-file#%s\n", rule.Code)
			fmt.Fprintf(wOut, "\tCategory: %s\n\n", rule.Category)
		}
	case "lint":
		_ = flagSet.Parse(os.Args[2:])
		if flagSet.NArg() < 1 {
			flagSet.Usage()
			os.Exit(2)
		}

		var configpath *string
		if *config != "" {
			configpath = config
		}

		// Multi args to allow usage where the shell expands wildcards like: ./pgvet migrations/*.sql
		patterns := flagSet.Args()[0:]

		os.Exit(lint(wOut, wErr, patterns, configpath, *format, *exitStatusOnViolations))
	default:
		flagSet.Usage()
		os.Exit(2)
	}
}

func lint(
	wOut, wErr io.Writer,
	patterns []string,
	configpath *string,
	format string,
	exitStatusOnViolations bool,
) int {
	log := newLogger(wErr)

	switch format {
	case formatJson, formatText:
	default:
		log.Error("Unknown format %q", format)
		return 1
	}

	cfg := defaultConfig()
	if configpath != nil {
		var err error
		cfg, err = overlayConfig(cfg, *configpath)
		if err != nil {
			log.Error("Failed to parse config: %s", err.Error())
			return 1
		}
	}

	fileMap := map[string]struct{}{}
	for _, pattern := range patterns {
		fileInfo, err := os.Stat(pattern)
		// If pattern is a directory be nice and parse all the files
		if err == nil && fileInfo.IsDir() {
			pattern = filepath.Join(pattern, "*")
		}
		patternFiles, err := filepath.Glob(pattern)
		if err != nil {
			log.Error(err.Error())
			return 1
		}
		for _, f := range patternFiles {
			i, err := os.Stat(f)
			if err != nil {
				continue
			}
			if !i.IsDir() {
				fileMap[f] = struct{}{}
			}
		}
	}

	if len(fileMap) == 0 {
		log.Error("No files found for patterns: %v", patterns)
		return 1
	}
	log.Info("Linting %d file(s)...\n\n", len(fileMap))

	var report Report
	for _, f := range slices.Sorted(maps.Keys(fileMap)) {
		content, err := os.ReadFile(f)
		if err != nil {
			log.Error("Failed to read file %s", err.Error())
			return 1
		}

		query := string(content)

		tree, err := pgquery.Parse(query)
		if err != nil {
			log.Error("Failed to parse SQL from file %q: %s", f, err.Error())
			return 1
		}

		var results []rules.Result
		for _, rule := range rules.AllRules() {
			if cfg, ok := cfg.Rules[rule.Code]; !ok || !cfg.Enabled {
				continue
			}
			partial, err := rule.Fn(tree, rule.Code, rule.Slug, rule.Help, *cfg.ImplicitTransaction)
			if err != nil {
				log.Error("Rule %q failed on file %q: %s", rule.Code, f, err.Error())
				return 1
			}
			results = append(results, partial...)
		}

		slices.SortFunc(results, func(a, b rules.Result) int {
			return cmp.Compare(a.StmtStart, b.StmtStart)
		})

		filtered := filterNoLints(query, results)
		for _, res := range filtered {
			statementLine := countLines(query[:res.StmtStart], query[res.StmtStart:res.StmtEnd])
			stmt := strings.TrimSpace(query[res.StmtStart:res.StmtEnd])
			entry := violation{
				File:          f,
				Code:          res.Code,
				Statement:     stmt,
				StatementLine: statementLine,
				Slug:          res.Slug,
				Help:          res.Help,
			}
			report = append(report, entry)
		}
	}

	serialized, err := report.Serialize(format)
	if err != nil {
		log.Error("Failed to seralize report: %s", err.Error())
		return 1
	}

	fmt.Fprint(wOut, serialized)

	if len(report) > 0 && exitStatusOnViolations {
		return 1
	}
	return 0
}

func countLines(precedingContent string, content string) int {
	precedingNumLines := len(strings.Split(strings.ReplaceAll(precedingContent, "\r\n", "\n"), "\n"))

	// The statement can contain newlines too which will be trimmed, so count them now
	var numLines int
	for _, line := range strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n") {
		if line == "" {
			numLines += 1
			continue
		}
		break
	}

	return precedingNumLines + numLines
}
