package analysis

// Level is review priority, not confidence or correctness.
type Level string

const (
	Low    Level = "low"
	Medium Level = "medium"
	High   Level = "high"
)

type Finding struct {
	Rule    string `json:"rule"`
	Level   Level  `json:"level"`
	Path    string `json:"path"`
	Line    int    `json:"line"`
	Comment string `json:"comment"`
}

const (
	Checked = "checked"
	Partial = "partial"
	Skipped = "skipped"
)

type Coverage struct {
	Path   string `json:"path"`
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
	Line   int    `json:"line,omitempty"`
}

type RunError struct {
	Path    string `json:"path,omitempty"`
	Line    int    `json:"line,omitempty"`
	Message string `json:"message"`
}

type FileResult struct {
	Findings []Finding
	Coverage Coverage
	Errors   []RunError
}

type Report struct {
	Status   string     `json:"status"`
	Findings []Finding  `json:"findings"`
	Coverage []Coverage `json:"coverage"`
	Errors   []RunError `json:"errors"`
}
