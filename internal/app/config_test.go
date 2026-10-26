package app

import (
	"os"
	"path/filepath"
	"testing"

	"pleat/internal/analysis"
)

func TestConfigurationChangesObservableFindings(t *testing.T) {
	dir := t.TempDir()
	source := writeInput(t, dir, "sample.py", "def sample():\n a = 1\n b = 2\n return a + b\n")
	config := writeInput(t, dir, "custom.toml", "[rules.\"layout.max-statement-run\"]\nmax = 2\nlevel = \"high\"\n[ rules.\"layout.return-separation\" ]\nenabled = false\n")
	report, code, _ := execute(t, "check", source, "--config", config, "--format=json")
	if code != 0 || report.Status != "complete" || len(report.Findings) != 1 {
		t.Fatalf("code=%d report=%+v", code, report)
	}
	finding := report.Findings[0]
	if finding.Rule != analysis.MaxStatementRun || finding.Level != analysis.High || finding.Line != 2 {
		t.Fatalf("configured finding=%+v", finding)
	}
}

func TestInvalidConfigurationIsAnError(t *testing.T) {
	for _, fixture := range []struct{ name, body string }{
		{"unknown-rule", "[rules.\"layout.not-a-rule\"]\nenabled = true\n"},
		{"unknown-key", "[rules.\"layout.max-statement-run\"]\nmaximum = 2\n"},
		{"bad-level", "[rules.\"layout.max-statement-run\"]\nlevel = \"critical\"\n"},
		{"zero-limit", "[rules.\"layout.max-statement-run\"]\nmax = 0\n"},
		{"wrong-rule-option", "[rules.\"layout.control-cuddle\"]\nmax = 2\n"},
		{"negative-block", "[rules.\"layout.return-separation\"]\nblock_limit = -1\n"},
		{"wrong-type", "[rules.\"layout.max-statement-run\"]\nenabled = \"yes\"\n"},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, ".pleat.toml"), []byte(fixture.body), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := loadSettings(dir, ""); err == nil {
				t.Fatalf("accepted invalid configuration %s", fixture.body)
			}
		})
	}
}

func TestConfigurationDoesNotSearchParentDirectories(t *testing.T) {
	parent := t.TempDir()
	child := filepath.Join(parent, "child")
	if err := os.Mkdir(child, 0700); err != nil {
		t.Fatal(err)
	}
	writeInput(t, parent, ".pleat.toml", "this is not TOML")
	if _, err := loadSettings(child, ""); err != nil {
		t.Fatalf("unexpected inherited configuration: %v", err)
	}
	if _, err := loadSettings(child, "missing.toml"); err == nil {
		t.Fatal("explicit missing configuration was ignored")
	}
}
