package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"pleat/internal/analysis"
)

const usage = `Pleat checks code layout with advisory findings.

Usage:
  pleat check FILE... [--config FILE] [--format text|json]
  pleat check --diff BASE [--config FILE] [--format text|json]

Files may be Go (.go), Python 3 (.py), or plain TypeScript (.ts).
--diff checks complete tracked working-tree files against the named commit.
Findings never fail the command. Operational errors exit 2.
`

type options struct {
	files  []string
	diff   string
	format string
	config string
	help   bool
}

// Run executes one invocation without changing the process working directory.
func Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	opts, err := parseCommand(args)
	if opts.help && err == nil {
		if _, err := io.WriteString(stdout, usage); err != nil {
			fmt.Fprintf(stderr, "pleat: write help: %v\n", err)
			return 2
		}
		return 0
	}
	report := emptyReport()
	if err != nil {
		report.Status = "failed"
		report.Errors = append(report.Errors, analysis.RunError{Message: err.Error()})
	} else {
		cwd, cwdErr := os.Getwd()
		if cwdErr != nil {
			report.Status = "failed"
			report.Errors = append(report.Errors, analysis.RunError{Message: "working directory: " + cwdErr.Error()})
		} else {
			report = run(ctx, cwd, opts)
		}
	}
	if err := render(stdout, opts.format, report); err != nil {
		fmt.Fprintf(stderr, "pleat: write report: %v\n", err)
		return 2
	}
	if len(report.Errors) > 0 {
		return 2
	}
	return 0
}

func parseCommand(args []string) (options, error) {
	opts := options{format: "text"}
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		opts.help = true
		return opts, nil
	}
	if len(args) == 0 || args[0] != "check" {
		return opts, fmt.Errorf("expected 'check'. Run 'pleat --help' for usage")
	}
	var firstError error
	seen := make(map[string]bool)
	literal := false
	for i := 1; i < len(args); i++ {
		arg := args[i]
		if literal || !strings.HasPrefix(arg, "-") || arg == "-" {
			opts.files = append(opts.files, arg)
			continue
		}
		if arg == "--" {
			literal = true
			continue
		}
		if arg == "--help" || arg == "-h" {
			opts.help = true
			continue
		}
		name, value, hasValue := strings.Cut(arg, "=")
		if name != "--diff" && name != "--format" && name != "--config" {
			if firstError == nil {
				firstError = fmt.Errorf("unknown option %q", name)
			}
			continue
		}
		if seen[name] && firstError == nil {
			firstError = fmt.Errorf("option %s supplied more than once", name)
		}
		seen[name] = true
		if !hasValue {
			if i+1 >= len(args) || strings.HasPrefix(args[i+1], "--") {
				if firstError == nil {
					firstError = fmt.Errorf("option %s requires a value", name)
				}
				continue
			}
			i++
			value = args[i]
		}
		if value == "" && firstError == nil {
			firstError = fmt.Errorf("option %s requires a nonempty value", name)
		}
		switch name {
		case "--diff":
			opts.diff = value
		case "--format":
			opts.format = value
		case "--config":
			opts.config = value
		}
	}
	if firstError != nil {
		return opts, firstError
	}
	if opts.format != "text" && opts.format != "json" {
		return opts, fmt.Errorf("unsupported format %q (use text or json)", opts.format)
	}
	if opts.help {
		return opts, nil
	}
	if opts.diff != "" && len(opts.files) != 0 {
		return opts, fmt.Errorf("explicit files and --diff are mutually exclusive")
	}
	if opts.diff == "" && len(opts.files) == 0 {
		return opts, fmt.Errorf("select files or use --diff BASE")
	}
	return opts, nil
}
