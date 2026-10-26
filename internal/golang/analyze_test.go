package golang

import (
	"context"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"pleat/internal/analysis"
)

type observedFinding struct {
	Rule  string
	Level analysis.Level
	Line  int
}

func settings() analysis.Settings {
	return analysis.Settings{
		StatementRun:     analysis.RuleSettings{Enabled: true, Level: analysis.Medium},
		ControlCuddle:    analysis.RuleSettings{Enabled: true, Level: analysis.Medium},
		ReturnSeparation: analysis.RuleSettings{Enabled: true, Level: analysis.Low},
		MaxStatements:    6,
		ReturnBlockLimit: 2,
	}
}

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	source, err := os.ReadFile("testdata/" + name + ".go")
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func expectFindings(t *testing.T, result analysis.FileResult, path string, want []observedFinding) {
	t.Helper()
	var got []observedFinding
	for _, finding := range result.Findings {
		if finding.Path != path {
			t.Errorf("finding path = %q, want %q", finding.Path, path)
		}
		got = append(got, observedFinding{finding.Rule, finding.Level, finding.Line})
	}
	sort.Slice(got, func(i, j int) bool {
		if got[i].Line != got[j].Line {
			return got[i].Line < got[j].Line
		}
		return got[i].Rule < got[j].Rule
	})
	if !reflect.DeepEqual(got, want) {
		t.Errorf("findings = %#v, want %#v", got, want)
	}
}

func expectChecked(t *testing.T, result analysis.FileResult, path string) {
	t.Helper()
	if result.Coverage != (analysis.Coverage{Path: path, Status: analysis.Checked}) {
		t.Errorf("coverage = %#v, want checked coverage for %q", result.Coverage, path)
	}
	if len(result.Errors) != 0 {
		t.Errorf("unexpected analysis errors: %#v", result.Errors)
	}
}

func TestDenseAndGroupedOrders(t *testing.T) {
	for _, tc := range []struct {
		name string
		want []observedFinding
	}{
		{"dense", []observedFinding{
			{analysis.MaxStatementRun, analysis.Medium, 4},
			{analysis.ReturnSeparation, analysis.Low, 12},
		}},
		{"grouped", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := "./reported/" + tc.name + ".go"
			result := Analyze(context.Background(), fixture(t, tc.name), path, settings())
			expectChecked(t, result, path)
			expectFindings(t, result, path, tc.want)
		})
	}
}

func TestEdgesKeepSmallBlocksAndConnectedClauses(t *testing.T) {
	config := settings()
	config.MaxStatements = 2
	result := Analyze(context.Background(), fixture(t, "edges"), "edges.go", config)
	expectChecked(t, result, "edges.go")
	expectFindings(t, result, "edges.go", []observedFinding{
		{analysis.MaxStatementRun, analysis.Medium, 27},
		{analysis.MaxStatementRun, analysis.Medium, 47},
		{analysis.ReturnSeparation, analysis.Low, 49},
	})
}

func TestStatementRunsUsePhysicalGapsBetweenStatements(t *testing.T) {
	source := []byte("package p\nfunc f() {\n first()\n // A comment alone does not separate a run.\n second()\n third()\n\n fourth()\n fifth()\n sixth()\n}\n")
	config := settings()
	config.MaxStatements = 2
	result := Analyze(context.Background(), source, "gaps.go", config)
	expectChecked(t, result, "gaps.go")
	expectFindings(t, result, "gaps.go", []observedFinding{
		{analysis.MaxStatementRun, analysis.Medium, 3},
		{analysis.MaxStatementRun, analysis.Medium, 8},
	})
}

func TestControlCuddleAssignmentAndHeaderMatching(t *testing.T) {
	cases := []struct {
		name string
		body string
		line int
	}{
		{"assignment and condition", "ready := check()\nif ready {}", 0},
		{"multiassignment and error", "value, err := read()\nif err != nil {}", 0},
		{"var declaration", "var ready = check()\nif ready {}", 0},
		{"compound assignment", "count += 1\nfor count < 4 {}", 0},
		{"range expression", "items := read()\nfor _, item := range items {}", 0},
		{"for header", "limit := read()\nfor i := 0; i < limit; i++ {}", 0},
		{"comment attached to condition", "ready := check()\n// Keep this with the guard.\nif ready {}", 0},
		{"first statement guard", "if ready {}", 0},
		{"unrelated assignment", "value := read()\nif ready {}", 4},
		{"two assignments", "first := read()\nready := check()\nif ready {}", 5},
		{"call before loop", "prepare()\nfor ready {}", 4},
		{"body is not header", "ready := check()\nif other { use(ready) }", 4},
		{"closure body is not header", "ready := check()\nif predicate(func() bool { return ready }) {}", 4},
		{"separated control", "prepare()\n\n// Guard comment.\nif ready {}", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			config := settings()
			config.StatementRun.Enabled = false
			config.ReturnSeparation.Enabled = false
			config.ControlCuddle.Level = analysis.High
			result := Analyze(context.Background(), []byte("package p\nfunc f() {\n"+tc.body+"\n}\n"), "control.go", config)
			expectChecked(t, result, "control.go")
			var want []observedFinding
			if tc.line != 0 {
				want = []observedFinding{{analysis.ControlCuddle, analysis.High, tc.line}}
			}
			expectFindings(t, result, "control.go", want)
		})
	}
}

func TestUncertainAssignmentRetainsIndependentRules(t *testing.T) {
	source := []byte("package p\nfunc f() {\n values[index] = ready\n if ready {}\n return\n}\n")
	config := settings()
	config.MaxStatements = 2
	result := Analyze(context.Background(), source, "uncertain.go", config)
	if result.Coverage.Path != "uncertain.go" || result.Coverage.Status != analysis.Partial || result.Coverage.Line != 4 || result.Coverage.Reason == "" {
		t.Errorf("coverage = %#v, want partial coverage at line 4", result.Coverage)
	}
	if len(result.Errors) != 0 {
		t.Errorf("uncertain layout classification is not a parser failure: %#v", result.Errors)
	}
	expectFindings(t, result, "uncertain.go", []observedFinding{
		{analysis.MaxStatementRun, analysis.Medium, 3},
		{analysis.ReturnSeparation, analysis.Low, 5},
	})
	config.ControlCuddle.Enabled = false
	result = Analyze(context.Background(), source, "uncertain.go", config)
	expectChecked(t, result, "uncertain.go")
	expectFindings(t, result, "uncertain.go", []observedFinding{
		{analysis.MaxStatementRun, analysis.Medium, 3},
		{analysis.ReturnSeparation, analysis.Low, 5},
	})
}

func TestMalformedDeclarationDoesNotLeakFindings(t *testing.T) {
	config := settings()
	config.MaxStatements = 2
	result := Analyze(context.Background(), fixture(t, "partial"), "partial.go", config)
	if result.Coverage.Status != analysis.Partial || result.Coverage.Path != "partial.go" || result.Coverage.Line != 12 || result.Coverage.Reason == "" {
		t.Errorf("coverage = %#v, want partial coverage at line 12", result.Coverage)
	}
	if len(result.Errors) == 0 || result.Errors[0].Path != "partial.go" || result.Errors[0].Line != 12 || result.Errors[0].Message == "" {
		t.Errorf("errors = %#v, want a located parser error at line 12", result.Errors)
	}
	expectFindings(t, result, "partial.go", []observedFinding{
		{analysis.MaxStatementRun, analysis.Medium, 4},
		{analysis.ReturnSeparation, analysis.Low, 6},
	})
}

func TestMalformedEnclosingBlockIsNotAnIndependentRegion(t *testing.T) {
	source := []byte("package p\nimport \"fmt\"\nfunc damaged() {\n first()\n second()\n if ready {\n  one()\n  two()\n  three()\n }\n broken :=\n}\n")
	config := settings()
	config.MaxStatements = 2
	result := Analyze(context.Background(), source, "broken.go", config)
	if result.Coverage.Status != analysis.Skipped || result.Coverage.Reason == "" || len(result.Errors) == 0 {
		t.Errorf("result = %#v, want skipped coverage and parser errors", result)
	}
	expectFindings(t, result, "broken.go", nil)
}

func TestLimitsLevelsAndDisabledRules(t *testing.T) {
	source := []byte("package p\nfunc f() {\n first()\n if ready {}\n return\n}\n")
	cases := []struct {
		name   string
		config analysis.Settings
		want   []observedFinding
	}{
		{"density only", analysis.Settings{
			StatementRun: analysis.RuleSettings{Enabled: true, Level: analysis.High}, MaxStatements: 2,
		}, []observedFinding{{analysis.MaxStatementRun, analysis.High, 3}}},
		{"return only", analysis.Settings{
			ReturnSeparation: analysis.RuleSettings{Enabled: true, Level: analysis.Medium}, ReturnBlockLimit: 1,
		}, []observedFinding{{analysis.ReturnSeparation, analysis.Medium, 5}}},
		{"control only", analysis.Settings{
			ControlCuddle: analysis.RuleSettings{Enabled: true, Level: analysis.Low},
		}, []observedFinding{{analysis.ControlCuddle, analysis.Low, 4}}},
		{"all disabled", analysis.Settings{}, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := Analyze(context.Background(), source, "settings.go", tc.config)
			expectChecked(t, result, "settings.go")
			expectFindings(t, result, "settings.go", tc.want)
		})
	}
}

func TestNestedFunctionsHaveIndependentRuns(t *testing.T) {
	source := []byte("package p\nfunc f() {\n callback := func() {\n  first()\n  second()\n  third()\n }\n use(callback)\n}\n")
	config := settings()
	config.MaxStatements = 2
	result := Analyze(context.Background(), source, "nested.go", config)
	expectChecked(t, result, "nested.go")
	expectFindings(t, result, "nested.go", []observedFinding{{analysis.MaxStatementRun, analysis.Medium, 4}})
}

func TestPhysicalLinesIgnoreLineDirectivesAndCRLF(t *testing.T) {
	source := []byte("package p\n//line generated.go:500\nfunc f() {\n first()\n second()\n return\n}\n")
	config := settings()
	config.MaxStatements = 2
	for _, source := range [][]byte{source, []byte(strings.ReplaceAll(string(source), "\n", "\r\n"))} {
		result := Analyze(context.Background(), source, "physical.go", config)
		expectChecked(t, result, "physical.go")
		expectFindings(t, result, "physical.go", []observedFinding{
			{analysis.MaxStatementRun, analysis.Medium, 4},
			{analysis.ReturnSeparation, analysis.Low, 6},
		})
	}
}

func TestCanceledAnalysisDoesNotClaimCheckedCoverage(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result := Analyze(ctx, fixture(t, "dense"), "canceled.go", settings())
	if result.Coverage.Status != analysis.Skipped || result.Coverage.Path != "canceled.go" || result.Coverage.Reason == "" {
		t.Errorf("coverage = %#v, want canceled skipped coverage", result.Coverage)
	}
	if len(result.Errors) != 1 || result.Errors[0].Path != "canceled.go" || result.Errors[0].Message != context.Canceled.Error() {
		t.Errorf("errors = %#v, want context cancellation", result.Errors)
	}
	expectFindings(t, result, "canceled.go", nil)
}

func TestMissingClosingBraceDoesNotProduceLayoutFindings(t *testing.T) {
	source := []byte("package p\nfunc damaged() {\n first()\n second()\n third()\n")
	config := settings()
	config.MaxStatements = 2
	result := Analyze(context.Background(), source, "unfinished.go", config)
	if result.Coverage.Status != analysis.Skipped || result.Coverage.Reason == "" || len(result.Errors) == 0 {
		t.Errorf("result = %#v, want skipped coverage and parser errors", result)
	}
	expectFindings(t, result, "unfinished.go", nil)
}

func TestSelectAndTypeSwitchClausesHaveIndependentSiblingBlocks(t *testing.T) {
	for _, source := range []string{
		"package p\nfunc f() {\n select {\n case value := <-ch:\n  first(value)\n  second()\n  return\n default:\n  fallback()\n }\n}\n",
		"package p\nfunc f() {\n switch value := input.(type) {\n case int:\n  first(value)\n  second()\n  return\n default:\n  fallback()\n }\n}\n",
	} {
		config := settings()
		config.MaxStatements = 2
		result := Analyze(context.Background(), []byte(source), "clauses.go", config)
		expectChecked(t, result, "clauses.go")
		expectFindings(t, result, "clauses.go", []observedFinding{
			{analysis.MaxStatementRun, analysis.Medium, 5},
			{analysis.ReturnSeparation, analysis.Low, 7},
		})
	}
}

func TestBlankLineInsideBlockCommentDoesNotSeparateRun(t *testing.T) {
	source := []byte("package p\nfunc f() {\n first()\n /* attached comment\n\n still a comment */\n second()\n third()\n}\n")
	config := settings()
	config.MaxStatements = 2
	result := Analyze(context.Background(), source, "comments.go", config)
	expectChecked(t, result, "comments.go")
	expectFindings(t, result, "comments.go", []observedFinding{{analysis.MaxStatementRun, analysis.Medium, 3}})
}
