package analysis

const (
	MaxStatementRun  = "layout.max-statement-run"
	ControlCuddle    = "layout.control-cuddle"
	ReturnSeparation = "layout.return-separation"
)

type RuleSettings struct {
	Enabled bool
	Level   Level
}

// Settings is validated once by the application before calling an analyzer.
type Settings struct {
	StatementRun     RuleSettings
	ControlCuddle    RuleSettings
	ReturnSeparation RuleSettings
	MaxStatements    int
	ReturnBlockLimit int
}
