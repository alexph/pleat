package app

import (
	"context"
	"os"
	"path/filepath"
	"sort"

	"pleat/internal/analysis"
	"pleat/internal/golang"
	"pleat/internal/python"
	"pleat/internal/typescript"
)

func emptyReport() analysis.Report {
	return analysis.Report{
		Findings: []analysis.Finding{},
		Coverage: []analysis.Coverage{},
		Errors:   []analysis.RunError{},
	}
}

func run(ctx context.Context, cwd string, opts options) analysis.Report {
	report := emptyReport()
	settings, err := loadSettings(cwd, opts.config)
	if err != nil {
		report.Status = "failed"
		report.Errors = append(report.Errors, analysis.RunError{Message: err.Error()})
		return report
	}
	selected, err := selectFiles(ctx, cwd, opts)
	if err != nil {
		report.Status = "failed"
		report.Errors = append(report.Errors, analysis.RunError{Message: err.Error()})
		return report
	}
	report.Coverage = append(report.Coverage, selected.coverage...)
	analyzed := 0
	incomplete := false
	for _, file := range selected.files {
		var result analysis.FileResult
		if err := ctx.Err(); err != nil {
			result = analysis.FileResult{
				Coverage: analysis.Coverage{Path: file.path, Status: analysis.Skipped, Reason: err.Error()},
				Errors:   []analysis.RunError{{Path: file.path, Message: err.Error()}},
			}
		} else {
			source, err := os.ReadFile(file.absolute)
			if err != nil {
				result = analysis.FileResult{
					Coverage: analysis.Coverage{Path: file.path, Status: analysis.Skipped, Reason: "source could not be read"},
					Errors:   []analysis.RunError{{Path: file.path, Message: err.Error()}},
				}
			} else {
				switch filepath.Ext(file.path) {
				case ".go":
					result = golang.Analyze(ctx, source, file.path, settings)
				case ".py":
					result = python.Analyze(ctx, source, file.path, settings)
				case ".ts":
					result = typescript.Analyze(ctx, source, file.path, settings)
				}
			}
		}
		report.Findings = append(report.Findings, result.Findings...)
		report.Coverage = append(report.Coverage, result.Coverage)
		report.Errors = append(report.Errors, result.Errors...)
		if result.Coverage.Status == analysis.Checked || result.Coverage.Status == analysis.Partial {
			analyzed++
		}
		if result.Coverage.Status != analysis.Checked {
			incomplete = true
		}
	}
	switch {
	case len(report.Errors) > 0 && analyzed == 0:
		report.Status = "failed"
	case incomplete || len(report.Errors) > 0:
		report.Status = "partial"
	case len(selected.files) == 0:
		report.Status = "empty"
	default:
		report.Status = "complete"
	}
	sort.Slice(report.Findings, func(i, j int) bool {
		left, right := report.Findings[i], report.Findings[j]
		if left.Path != right.Path {
			return left.Path < right.Path
		}
		if left.Line != right.Line {
			return left.Line < right.Line
		}
		if left.Rule != right.Rule {
			return left.Rule < right.Rule
		}
		if left.Level != right.Level {
			return left.Level < right.Level
		}
		return left.Comment < right.Comment
	})
	sort.Slice(report.Coverage, func(i, j int) bool {
		left, right := report.Coverage[i], report.Coverage[j]
		if left.Path != right.Path {
			return left.Path < right.Path
		}
		if left.Line != right.Line {
			return left.Line < right.Line
		}
		if left.Status != right.Status {
			return left.Status < right.Status
		}
		return left.Reason < right.Reason
	})
	sort.Slice(report.Errors, func(i, j int) bool {
		left, right := report.Errors[i], report.Errors[j]
		if left.Path != right.Path {
			return left.Path < right.Path
		}
		if left.Line != right.Line {
			return left.Line < right.Line
		}
		return left.Message < right.Message
	})
	return report
}
