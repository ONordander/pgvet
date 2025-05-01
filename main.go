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

	"github.com/onordander/pgcheck/rules"

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

	flag.CommandLine.SetOutput(wErr)
	verbose := flag.Bool("verbose", false, "Set verbose debugging output")
	format := flag.String("format", "text", "Set output format, text or json. Text is default")
	config := flag.String("config", "", "Config file")
	flag.Usage = func() {
		fmt.Fprint(wErr, "Usage:\n")
		fmt.Fprint(wErr, "\t./pgcheck lint [-verbose] [-config <config.yaml>] <filepattern>...\n")
		fmt.Fprint(wErr, "\t./pgcheck --help\n")
		fmt.Fprint(wErr, "\t./pgcheck version\n")
		fmt.Fprint(wErr, "\t./pgcheck license\n")
		flag.PrintDefaults()
		fmt.Fprint(wErr, "Example:\n")
		fmt.Fprint(wErr, "\t./pgcheck lint -config config.yaml migrations/*.sql\n")
	}
	flag.Parse()

	switch flag.Arg(0) {
	case "license":
		fmt.Fprint(wOut, notice)
		return
	case "version":
		fmt.Fprintf(wOut, "pgcheck %s\n", strings.TrimSpace(version))
		return
	case "rules":
		for _, rule := range rules.AllRules() {
			enabled := "✓"
			if rule.DisabledByDefault {
				enabled = "🗙"
			}
			fmt.Fprintf(wOut, "%s%s%s\n", magenta, rule.Code, normal)
			fmt.Fprintf(wOut, "\t| %s\n", rule.Slug)
			fmt.Fprintf(wOut, "\tHelp: %s\n", rule.Help)
			fmt.Fprintf(wOut, "\tEnabled by default: %s\n", enabled)
			fmt.Fprintf(wOut, "\tExplanation: https://github.com/ONordander/pgcheck?tab=readme-ov-file#%s\n", rule.Code)
			fmt.Fprintf(wOut, "\tCategory: %s\n\n", rule.Category)
		}
	case "lint":
		if flag.NArg() < 2 {
			flag.Usage()
			os.Exit(2)
		}

		var configpath *string
		if *config != "" {
			configpath = config
		}

		// Multi args to allow usage where the shell expands wildcards like: ./pgcheck migrations/*.sql
		patterns := flag.Args()[1:]

		os.Exit(lint(wOut, wErr, *verbose, patterns, configpath, *format))
	default:
		flag.Usage()
		os.Exit(2)
	}
}

func lint(
	wOut, wErr io.Writer,
	verbose bool,
	patterns []string,
	configpath *string,
	format string,
) int {
	logger := configureLogger(verbose, wErr)

	switch format {
	case formatJson, formatText:
	default:
		logger.Error(fmt.Sprintf("Unknown format %q", format))
		return 1
	}

	cfg := defaultConfig()
	if configpath != nil {
		var err error
		cfg, err = overlayConfig(cfg, *configpath)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to parse config: %s", err.Error()))
			return 1
		}
	}

	fileMap := map[string]struct{}{}
	for _, pattern := range patterns {
		patternFiles, err := filepath.Glob(pattern)
		if err != nil {
			logger.Error(err.Error())
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
		logger.Error("No files found")
		return 1
	}

	var report Report
	for _, f := range slices.Sorted(maps.Keys(fileMap)) {
		content, err := os.ReadFile(f)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to read file %s", err.Error()))
			return 1
		}

		query := string(content)

		tree, err := pgquery.Parse(query)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to parse file %q: %s", f, err.Error()))
			return 1
		}

		var results []rules.Result
		for _, rule := range rules.AllRules() {
			if cfg, ok := cfg.Rules[rule.Code]; !ok || !cfg.Enabled {
				continue
			}
			partial, err := rule.Fn(tree, rule.Code, rule.Slug, rule.Help)
			if err != nil {
				logger.Error(fmt.Sprintf("Rule %q failed on file %q: %s", rule.Code, f, err.Error()))
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
			entry := lintError{
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
		logger.Error(fmt.Sprintf("Failed to seralize report: %s", err.Error()))
		return 1
	}

	fmt.Fprint(wOut, serialized)
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
