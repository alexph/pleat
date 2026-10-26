package app

import (
	"fmt"
	"io"
	"strconv"

	"pleat/internal/analysis"
)

func render(writer io.Writer, format string, report analysis.Report) error {
	if format == "json" {
		return writeJSON(writer, report)
	}
	return writeText(writer, report)
}

func writeText(writer io.Writer, report analysis.Report) error {
	if _, err := fmt.Fprintf(writer, "Pleat: %s (%d findings)\n", report.Status, len(report.Findings)); err != nil {
		return err
	}
	for _, finding := range report.Findings {
		if _, err := fmt.Fprintf(writer, "%s:%d [%s] %s: %s\n", strconv.Quote(finding.Path), finding.Line, finding.Level, finding.Rule, finding.Comment); err != nil {
			return err
		}
	}
	for _, coverage := range report.Coverage {
		location := strconv.Quote(coverage.Path)
		if coverage.Line > 0 {
			location += fmt.Sprintf(":%d", coverage.Line)
		}
		if _, err := fmt.Fprintf(writer, "Coverage %s: %s", location, coverage.Status); err != nil {
			return err
		}
		if coverage.Reason != "" {
			if _, err := fmt.Fprintf(writer, " (%s)", strconv.Quote(coverage.Reason)); err != nil {
				return err
			}
		}
		if _, err := io.WriteString(writer, "\n"); err != nil {
			return err
		}
	}
	for _, runError := range report.Errors {
		location := ""
		if runError.Path != "" {
			location = strconv.Quote(runError.Path)
			if runError.Line > 0 {
				location += fmt.Sprintf(":%d", runError.Line)
			}
			location += ": "
		}
		if _, err := fmt.Fprintf(writer, "Error %s%s\n", location, strconv.Quote(runError.Message)); err != nil {
			return err
		}
	}
	return nil
}
