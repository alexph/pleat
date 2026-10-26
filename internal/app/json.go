package app

import (
	"encoding/json"
	"io"

	"pleat/internal/analysis"
)

func writeJSON(writer io.Writer, report analysis.Report) error {
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}
