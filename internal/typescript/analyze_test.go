package typescript

import (
	"context"
	"os"
	"reflect"
	"sort"
	"testing"

	"pleat/internal/analysis"
)

func testSettings() analysis.Settings {
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
	source, err := os.ReadFile("testdata/" + name + ".ts")
	if err != nil {
		t.Fatal(err)
	}
	return source
}

type observation struct {
	Rule  string
	Line  int
	Level analysis.Level
}

func assertFindings(t *testing.T, result analysis.FileResult, path string, expected []observation) {
	t.Helper()
	var actual []observation
	for _, finding := range result.Findings {
		if finding.Path != path {
			t.Errorf("finding path = %q, want %q", finding.Path, path)
		}
		actual = append(actual, observation{finding.Rule, finding.Line, finding.Level})
	}
	sort.Slice(actual, func(i, j int) bool {
		if actual[i].Line != actual[j].Line {
			return actual[i].Line < actual[j].Line
		}
		return actual[i].Rule < actual[j].Rule
	})
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("findings = %#v, want %#v", actual, expected)
	}
}

func assertChecked(t *testing.T, result analysis.FileResult, path string) {
	t.Helper()
	if result.Coverage != (analysis.Coverage{Path: path, Status: analysis.Checked}) {
		t.Errorf("coverage = %#v, want checked for %q", result.Coverage, path)
	}
	if len(result.Errors) != 0 {
		t.Errorf("unexpected analysis errors: %#v", result.Errors)
	}
}

func TestOrderFixtures(t *testing.T) {
	for _, tc := range []struct {
		name string
		want []observation
	}{
		{"dense", []observation{{"layout.max-statement-run", 2, analysis.Medium}, {"layout.return-separation", 10, analysis.Low}}},
		{"grouped", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := "reported/" + tc.name + ".ts"
			result := Analyze(context.Background(), fixture(t, tc.name), path, testSettings())
			assertChecked(t, result, path)
			assertFindings(t, result, path, tc.want)
		})
	}
}

func TestEdges(t *testing.T) {
	settings := testSettings()
	settings.MaxStatements = 3
	result := Analyze(context.Background(), fixture(t, "edges"), "edges.ts", settings)
	assertChecked(t, result, "edges.ts")
	assertFindings(t, result, "edges.ts", []observation{
		{"layout.max-statement-run", 25, analysis.Medium},
		{"layout.max-statement-run", 46, analysis.Medium},
		{"layout.max-statement-run", 53, analysis.Medium},
		{"layout.return-separation", 56, analysis.Low},
		{"layout.control-cuddle", 65, analysis.Medium},
		{"layout.control-cuddle", 73, analysis.Medium},
	})
}

func TestRuleSettings(t *testing.T) {
	source := []byte("function check() {\n  prepare();\n  if (ready()) { run(); }\n  return result;\n}\n")
	for _, tc := range []struct {
		name     string
		settings analysis.Settings
		want     []observation
	}{
		{
			"all priorities",
			analysis.Settings{
				StatementRun:     analysis.RuleSettings{Enabled: true, Level: analysis.High},
				ControlCuddle:    analysis.RuleSettings{Enabled: true, Level: analysis.Low},
				ReturnSeparation: analysis.RuleSettings{Enabled: true, Level: analysis.Medium},
				MaxStatements:    2, ReturnBlockLimit: 2,
			},
			[]observation{{"layout.max-statement-run", 2, analysis.High}, {"layout.control-cuddle", 3, analysis.Low}, {"layout.return-separation", 4, analysis.Medium}},
		},
		{
			"density only",
			analysis.Settings{StatementRun: analysis.RuleSettings{Enabled: true, Level: analysis.Low}, MaxStatements: 2, ReturnBlockLimit: 2},
			[]observation{{"layout.max-statement-run", 2, analysis.Low}},
		},
		{
			"control only",
			analysis.Settings{ControlCuddle: analysis.RuleSettings{Enabled: true, Level: analysis.High}, MaxStatements: 2, ReturnBlockLimit: 2},
			[]observation{{"layout.control-cuddle", 3, analysis.High}},
		},
		{
			"return only",
			analysis.Settings{ReturnSeparation: analysis.RuleSettings{Enabled: true, Level: analysis.High}, MaxStatements: 2, ReturnBlockLimit: 2},
			[]observation{{"layout.return-separation", 4, analysis.High}},
		},
		{
			"raised thresholds",
			analysis.Settings{StatementRun: analysis.RuleSettings{Enabled: true, Level: analysis.High}, ReturnSeparation: analysis.RuleSettings{Enabled: true, Level: analysis.High}, MaxStatements: 3, ReturnBlockLimit: 3},
			nil,
		},
		{"disabled", analysis.Settings{MaxStatements: 1, ReturnBlockLimit: 1}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := Analyze(context.Background(), source, "settings.ts", tc.settings)
			assertChecked(t, result, "settings.ts")
			assertFindings(t, result, "settings.ts", tc.want)
		})
	}
}

func TestMaximalRunsAndMultilineStatements(t *testing.T) {
	source := []byte("function check() {\n  first(); second();\n  third();\n\n  const value = transform<Result>(\n    first(),\n\n    second(),\n  );\n  use(value);\n  return value;\n}\n")
	settings := testSettings()
	settings.MaxStatements = 2
	result := Analyze(context.Background(), source, "runs.ts", settings)
	assertChecked(t, result, "runs.ts")
	assertFindings(t, result, "runs.ts", []observation{
		{"layout.max-statement-run", 2, analysis.Medium},
		{"layout.max-statement-run", 5, analysis.Medium},
		{"layout.return-separation", 11, analysis.Low},
	})
}

func TestAssignmentHeaders(t *testing.T) {
	for _, tc := range []struct {
		name       string
		assignment string
		header     string
		want       []observation
	}{
		{"assignment", "ready = read();", "if (ready) { run(); }", nil},
		{"augmented assignment", "ready ||= read();", "while (ready) { break; }", nil},
		{"destructuring", "const [ready, other] = read();", "if (ready) { run(); }", nil},
		{"renamed binding", "const {ready: active} = read();", "if (active) { run(); }", nil},
		{"for initializer", "let limit = read();", "for (let i = 0; i < limit; i++) { run(); }", nil},
		{"do condition", "ready = read();", "do { run(); } while (ready);", nil},
		{"unrelated name", "const ready = read();", "if (other) { run(); }", []observation{{"layout.control-cuddle", 3, analysis.Medium}}},
		{"property is not a reference", "const ready = read();", "if (object.ready) { run(); }", []observation{{"layout.control-cuddle", 3, analysis.Medium}}},
		{"uninitialized declaration", "let ready;", "if (ready) { run(); }", []observation{{"layout.control-cuddle", 3, analysis.Medium}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := []byte("function check() {\n  " + tc.assignment + "\n  " + tc.header + "\n}\n")
			result := Analyze(context.Background(), source, "pairs.ts", testSettings())
			assertChecked(t, result, "pairs.ts")
			assertFindings(t, result, "pairs.ts", tc.want)
		})
	}
}

func TestUnknownAssignmentKeepsOtherRules(t *testing.T) {
	source := []byte("function check() {\n  state[key] = read();\n  if (state) { run(); }\n  return state;\n}\n")
	settings := testSettings()
	settings.MaxStatements = 2
	result := Analyze(context.Background(), source, "unknown.ts", settings)
	if result.Coverage.Path != "unknown.ts" || result.Coverage.Status != analysis.Partial || result.Coverage.Line != 3 || result.Coverage.Reason == "" || len(result.Errors) != 0 {
		t.Errorf("unknown assignment coverage/errors = %#v / %#v", result.Coverage, result.Errors)
	}
	assertFindings(t, result, "unknown.ts", []observation{
		{"layout.max-statement-run", 2, analysis.Medium},
		{"layout.return-separation", 4, analysis.Low},
	})
	settings.ControlCuddle.Enabled = false
	result = Analyze(context.Background(), source, "unknown.ts", settings)
	assertChecked(t, result, "unknown.ts")
	assertFindings(t, result, "unknown.ts", []observation{
		{"layout.max-statement-run", 2, analysis.Medium},
		{"layout.return-separation", 4, analysis.Low},
	})
}

func TestPartialSyntaxRetainsIndependentFunction(t *testing.T) {
	result := Analyze(context.Background(), fixture(t, "partial"), "partial.ts", testSettings())
	if result.Coverage.Path != "partial.ts" || result.Coverage.Status != analysis.Partial || result.Coverage.Line == 0 || result.Coverage.Reason == "" {
		t.Errorf("coverage = %#v, want partial with a location and reason", result.Coverage)
	}
	if len(result.Errors) == 0 || result.Errors[0].Path != "partial.ts" || result.Errors[0].Line == 0 {
		t.Errorf("errors = %#v, want a located parser error", result.Errors)
	}
	assertFindings(t, result, "partial.ts", []observation{
		{"layout.max-statement-run", 2, analysis.Medium},
		{"layout.return-separation", 10, analysis.Low},
		{"layout.return-separation", 24, analysis.Low},
	})
}

func TestMalformedEnclosingRegionIsSkipped(t *testing.T) {
	for _, source := range []string{
		"function broken() {\n  first();\n  second();\n  return value;\n",
		"function broken() {\n  const value = ;\n  function inner() { first(); second(); return value; }\n}\n",
	} {
		settings := testSettings()
		settings.MaxStatements = 1
		result := Analyze(context.Background(), []byte(source), "broken.ts", settings)
		if result.Coverage.Status != analysis.Skipped || result.Coverage.Reason == "" || len(result.Errors) == 0 {
			t.Errorf("result = %#v, want skipped with parser error", result)
		}
		assertFindings(t, result, "broken.ts", nil)
	}
}

func TestCancelledAnalysis(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result := Analyze(ctx, fixture(t, "dense"), "cancelled.ts", testSettings())
	if result.Coverage.Status != analysis.Skipped || len(result.Errors) != 1 || result.Errors[0].Path != "cancelled.ts" {
		t.Errorf("result = %#v, want skipped with cancellation error", result)
	}
	assertFindings(t, result, "cancelled.ts", nil)
}

func TestTypeOnlyDeclarationsDoNotCount(t *testing.T) {
	source := []byte("import type {Input} from './input';\nexport type {Output} from './output';\nfirst();\ninterface Marker {}\ntype Alias = string;\ndeclare const external: string;\nsecond();\n\nfunction check() {\n  const value = read();\n  interface Local {}\n  type Result = Input;\n  return value;\n}\n")
	settings := testSettings()
	settings.MaxStatements = 2
	result := Analyze(context.Background(), source, "types.ts", settings)
	assertChecked(t, result, "types.ts")
	assertFindings(t, result, "types.ts", nil)
}
