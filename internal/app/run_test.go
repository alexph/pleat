package app

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"pleat/internal/analysis"
)

func execute(t *testing.T, args ...string) (analysis.Report, int, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := Run(context.Background(), args, &stdout, &stderr)
	var report analysis.Report
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("decode CLI JSON: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
	return report, code, stdout.String()
}

func writeInput(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestMixedLanguageAdvisoryReport(t *testing.T) {
	dir := t.TempDir()
	fixtureInput := func(name string) string {
		data, err := os.ReadFile(filepath.Join("testdata", name))
		if err != nil {
			t.Fatal(err)
		}
		return writeInput(t, dir, name, string(data))
	}
	goPath := fixtureInput("orders.go")
	pyPath := fixtureInput("orders.py")
	tsPath := fixtureInput("orders.ts")
	args := []string{"check", tsPath, pyPath, goPath, "--format", "json"}
	report, code, raw := execute(t, args...)
	if code != 0 || report.Status != "complete" || len(report.Errors) != 0 {
		t.Fatalf("unexpected result code=%d report=%+v", code, report)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	var expected []struct {
		path, rule string
		line       int
		level      analysis.Level
	}
	for _, row := range []struct {
		path       string
		start, end int
	}{{goPath, 4, 12}, {pyPath, 2, 10}, {tsPath, 2, 10}} {
		rel, err := filepath.Rel(cwd, row.path)
		if err != nil {
			t.Fatal(err)
		}
		expected = append(expected, struct {
			path, rule string
			line       int
			level      analysis.Level
		}{filepath.ToSlash(rel), analysis.MaxStatementRun, row.start, analysis.Medium}, struct {
			path, rule string
			line       int
			level      analysis.Level
		}{filepath.ToSlash(rel), analysis.ReturnSeparation, row.end, analysis.Low})
	}
	if len(report.Findings) != len(expected) {
		t.Fatalf("findings=%+v", report.Findings)
	}
	for i, want := range expected {
		got := report.Findings[i]
		if got.Path != want.path || got.Rule != want.rule || got.Line != want.line || got.Level != want.level {
			t.Fatalf("finding %d=%+v, want %+v", i, got, want)
		}
	}
	_, _, repeated := execute(t, args...)
	if raw != repeated {
		t.Fatal("unchanged input produced different ordered JSON")
	}
	for _, path := range []string{goPath, pyPath, tsPath} {
		before, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		_, _, _ = execute(t, "check", path, "--format=json")
		after, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(before, after) {
			t.Fatalf("analysis modified %s", path)
		}
	}
}

func TestParserFailureRetainsGoodFile(t *testing.T) {
	dir := t.TempDir()
	good := writeInput(t, dir, "good.py", "def good():\n a = 1\n b = 2\n return a + b\n")
	bad := writeInput(t, dir, "bad.py", "def broken(:\n")
	report, code, _ := execute(t, "check", bad, good, "--format", "json")
	if code != 2 || report.Status != "partial" || len(report.Errors) == 0 {
		t.Fatalf("code=%d report=%+v", code, report)
	}
	if len(report.Findings) != 1 || report.Findings[0].Rule != analysis.ReturnSeparation {
		t.Fatalf("good file findings lost: %+v", report.Findings)
	}
	if len(report.Coverage) != 2 || report.Coverage[0].Status != analysis.Skipped || report.Coverage[1].Status != analysis.Checked {
		t.Fatalf("coverage=%+v", report.Coverage)
	}
}

func TestUnreadableSourceAndInvalidInvocation(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.go")
	report, code, _ := execute(t, "check", missing, "--format=json")
	if code != 2 || report.Status != "failed" || len(report.Errors) != 1 || report.Coverage[0].Status != analysis.Skipped {
		t.Fatalf("code=%d report=%+v", code, report)
	}
	for _, args := range [][]string{{"check", "--format=json"}, {"check", "source.go", "--diff=HEAD", "--format=json"}, {"check", "unsupported.txt", "--format=json"}, {"check", "--unknown", "--format=json"}} {
		report, code, _ := execute(t, args...)
		if code != 2 || report.Status != "failed" || len(report.Errors) == 0 {
			t.Fatalf("args=%v code=%d report=%+v", args, code, report)
		}
		if !reflect.DeepEqual(report.Findings, []analysis.Finding{}) || !reflect.DeepEqual(report.Coverage, []analysis.Coverage{}) {
			t.Fatalf("invalid invocation analyzed source: %+v", report)
		}
	}
}

func TestEmptyDiffDoesNotClaimRepositoryPassed(t *testing.T) {
	dir := t.TempDir()
	gitTest(t, dir, "init", "--quiet")
	writeInput(t, dir, "source.go", "package source\n")
	writeInput(t, dir, "notes.txt", "baseline\n")
	gitTest(t, dir, "add", ".")
	gitTest(t, dir, "commit", "--quiet", "-m", "baseline")
	if err := os.Remove(filepath.Join(dir, "source.go")); err != nil {
		t.Fatal(err)
	}
	writeInput(t, dir, "notes.txt", "changed outside scope\n")
	report := run(context.Background(), dir, options{diff: "HEAD", format: "json"})
	if report.Status != "empty" || len(report.Findings) != 0 || len(report.Errors) != 0 || len(report.Coverage) != 2 {
		t.Fatalf("empty-scope report=%+v", report)
	}
	for _, coverage := range report.Coverage {
		if coverage.Status != analysis.Skipped || coverage.Reason == "" {
			t.Fatalf("coverage=%+v", coverage)
		}
	}
}

func TestCancelledInvocationReportsNoCleanAnalysis(t *testing.T) {
	dir := t.TempDir()
	path := writeInput(t, dir, "source.py", "def source():\n return 1\n")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var stdout, stderr bytes.Buffer
	code := Run(ctx, []string{"check", path, "--format=json"}, &stdout, &stderr)
	var report analysis.Report
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if code != 2 || report.Status != "failed" || len(report.Errors) != 1 || report.Coverage[0].Status != analysis.Skipped {
		t.Fatalf("cancelled code=%d report=%+v", code, report)
	}
}
