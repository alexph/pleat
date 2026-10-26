package python

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

func explicitSettings() analysis.Settings {
	return analysis.Settings{
		StatementRun:     analysis.RuleSettings{Enabled: true, Level: analysis.Medium},
		ControlCuddle:    analysis.RuleSettings{Enabled: true, Level: analysis.Medium},
		ReturnSeparation: analysis.RuleSettings{Enabled: true, Level: analysis.Low},
		MaxStatements:    6, ReturnBlockLimit: 2,
	}
}

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	source, err := os.ReadFile("testdata/" + name + ".py")
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func assertFindings(t *testing.T, result analysis.FileResult, want []observedFinding) {
	t.Helper()
	got := make([]observedFinding, 0, len(result.Findings))
	for _, finding := range result.Findings {
		if finding.Path != "reported/source.py" {
			t.Errorf("finding path = %q", finding.Path)
		}
		got = append(got, observedFinding{finding.Rule, finding.Level, finding.Line})
	}
	less := func(values []observedFinding, i, j int) bool {
		if values[i].Line != values[j].Line {
			return values[i].Line < values[j].Line
		}
		return values[i].Rule < values[j].Rule
	}
	sort.Slice(got, func(i, j int) bool { return less(got, i, j) })
	if want == nil {
		want = []observedFinding{}
	}
	sort.Slice(want, func(i, j int) bool { return less(want, i, j) })
	if !reflect.DeepEqual(got, want) {
		t.Errorf("findings = %#v, want %#v", got, want)
	}
}

func assertChecked(t *testing.T, result analysis.FileResult) {
	t.Helper()
	if result.Coverage.Path != "reported/source.py" || result.Coverage.Status != analysis.Checked || len(result.Errors) != 0 {
		t.Errorf("unexpected analysis coverage/errors: %#v / %#v", result.Coverage, result.Errors)
	}
}

func TestVisionFixtures(t *testing.T) {
	for _, name := range []string{"dense", "grouped", "edges"} {
		t.Run(name, func(t *testing.T) {
			settings := explicitSettings()
			if name != "dense" {
				settings.MaxStatements = 2
			}
			result := Analyze(context.Background(), fixture(t, name), "reported/source.py", settings)
			assertChecked(t, result)
			var want []observedFinding
			if name == "dense" {
				want = []observedFinding{{analysis.MaxStatementRun, analysis.Medium, 2}, {analysis.ReturnSeparation, analysis.Low, 10}}
			}
			assertFindings(t, result, want)
		})
	}
}

func TestStatementBoundaries(t *testing.T) {
	cases := []struct {
		name   string
		source string
		want   []observedFinding
	}{
		{"comment is not separation", "def f():\n    a()\n    # attached\n    b()\n    c()\n", []observedFinding{{analysis.MaxStatementRun, analysis.Medium, 2}}},
		{"whitespace CRLF separator", "def f():\r\n    a()\r\n \t\r\n    # attached\r\n    b()\r\n    c()\r\n", nil},
		{"internal blank does not separate", "def f():\n    a = call(\n        1,\n\n        2,\n    )\n    b()\n    c()\n", []observedFinding{{analysis.MaxStatementRun, analysis.Medium, 2}}},
		{"nested block independent", "def f():\n    if ready:\n        a()\n        b()\n        c()\n    done()\n", []observedFinding{{analysis.MaxStatementRun, analysis.Medium, 3}}},
		{"semicolon siblings", "def f():\n    a(); b(); c()\n", []observedFinding{{analysis.MaxStatementRun, analysis.Medium, 2}}},
		{"one finding per maximal run", "def f():\n    a()\n    b()\n    c()\n\n    d()\n    e()\n    f()\n", []observedFinding{{analysis.MaxStatementRun, analysis.Medium, 2}, {analysis.MaxStatementRun, analysis.Medium, 6}}},
		{"docstring is not executable", "def f():\n    \"\"\"Explanation.\"\"\"\n    answer = 42\n    return answer\n", nil},
		{"parenthesized docstring comments", "def f():\n    (\n        # attached\n        \"Explanation.\"\n    )\n    answer = 42\n    return answer\n", nil},
		{"concatenated docstring", "def f():\n    (\"First. \" \"Second.\")\n    answer = 42\n    return answer\n", nil},
		{"decorated statement counts once", "def outer():\n    @first\n    @second\n    def inner():\n        pass\n    done()\n", nil},
		{"module and class docstrings", "\"Module explanation.\"\nclass Example:\n    \"Class explanation.\"\n    first = 1\n    second = 2\n", nil},
		{"later string is executable", "def f():\n    answer = 42\n    \"not a docstring\"\n    return answer\n", []observedFinding{{analysis.MaxStatementRun, analysis.Medium, 2}, {analysis.ReturnSeparation, analysis.Low, 4}}},
		{"formatted string is executable", "def f():\n    f\"{value}\"\n    answer = 42\n    return answer\n", []observedFinding{{analysis.MaxStatementRun, analysis.Medium, 2}, {analysis.ReturnSeparation, analysis.Low, 4}}},
		{"case clauses stay connected", "def f(value):\n    match value:\n        case 1:\n            a()\n            b()\n            c()\n        case _:\n            return None\n", []observedFinding{{analysis.MaxStatementRun, analysis.Medium, 4}}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			settings := explicitSettings()
			settings.MaxStatements = 2
			result := Analyze(context.Background(), []byte(test.source), "reported/source.py", settings)
			assertChecked(t, result)
			assertFindings(t, result, test.want)
		})
	}
}

func TestControlAdjacency(t *testing.T) {
	cases := []struct {
		name   string
		source string
		want   []observedFinding
	}{
		{"assigned condition", "def f():\n    ready = check()\n    # attached\n    if ready:\n        return ready\n", nil},
		{"two statement preceding group", "def f():\n    prepare()\n    ready = check()\n    if ready:\n        pass\n", []observedFinding{{analysis.ControlCuddle, analysis.Medium, 3}}},
		{"unrelated assignment", "def f():\n    ready = check()\n    if other:\n        pass\n", []observedFinding{{analysis.ControlCuddle, analysis.Medium, 3}}},
		{"header does not include body", "def f():\n    ready = check()\n    if other:\n        consume(ready)\n", []observedFinding{{analysis.ControlCuddle, analysis.Medium, 3}}},
		{"exact identifier matching", "def f():\n    ready = check()\n    if already_ready:\n        pass\n", []observedFinding{{analysis.ControlCuddle, analysis.Medium, 3}}},
		{"destructuring assignment", "def f():\n    first, *rest = values()\n    if rest:\n        pass\n", nil},
		{"chained assignment", "def f():\n    first = second = check()\n    while second:\n        break\n", nil},
		{"augmented assignment", "def f():\n    count += 1\n    while count:\n        break\n", nil},
		{"annotated assignment", "def f():\n    ready: bool = check()\n    if ready:\n        pass\n", nil},
		{"annotation without value", "def f():\n    ready: bool\n    if ready:\n        pass\n", []observedFinding{{analysis.ControlCuddle, analysis.Medium, 3}}},
		{"for iterable", "async def f():\n    items = load()\n    async for item in items:\n        pass\n", nil},
		{"for target is syntactic header", "def f():\n    item = previous()\n    for item in values:\n        pass\n", nil},
		{"for body does not exempt", "def f():\n    ready = check()\n    for item in values:\n        consume(ready)\n", []observedFinding{{analysis.ControlCuddle, analysis.Medium, 3}}},
		{"blank before attached comment", "def f():\n    prepare()\n\n    # attached\n    if ready:\n        pass\n", nil},
		{"continuation clauses", "def f():\n    if ready:\n        pass\n    elif other:\n        pass\n    else:\n        pass\n", nil},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			result := Analyze(context.Background(), []byte(test.source), "reported/source.py", explicitSettings())
			assertChecked(t, result)
			assertFindings(t, result, test.want)
		})
	}
}

func TestUnsupportedAssignmentRetainsOtherRules(t *testing.T) {
	source := []byte("def f():\n    obj.ready = check()\n    if obj.ready:\n        pass\n    finish()\n    return obj\n")
	settings := explicitSettings()
	settings.MaxStatements = 2
	result := Analyze(context.Background(), source, "reported/source.py", settings)
	if result.Coverage.Status != analysis.Partial || result.Coverage.Line != 2 || result.Coverage.Reason == "" || len(result.Errors) != 0 {
		t.Errorf("unsupported coverage/errors = %#v / %#v", result.Coverage, result.Errors)
	}
	assertFindings(t, result, []observedFinding{{analysis.MaxStatementRun, analysis.Medium, 2}, {analysis.ReturnSeparation, analysis.Low, 6}})
	settings.ControlCuddle.Enabled = false
	result = Analyze(context.Background(), source, "reported/source.py", settings)
	assertChecked(t, result)
	assertFindings(t, result, []observedFinding{{analysis.MaxStatementRun, analysis.Medium, 2}, {analysis.ReturnSeparation, analysis.Low, 6}})
}

func TestDamagedRegionDoesNotProduceFindings(t *testing.T) {
	settings := explicitSettings()
	settings.MaxStatements = 2
	result := Analyze(context.Background(), fixture(t, "partial"), "reported/source.py", settings)
	if result.Coverage.Status != analysis.Partial || result.Coverage.Reason == "" || result.Coverage.Line < 1 || len(result.Errors) == 0 {
		t.Errorf("partial coverage/errors = %#v / %#v", result.Coverage, result.Errors)
	}
	assertFindings(t, result, []observedFinding{{analysis.MaxStatementRun, analysis.Medium, 9}, {analysis.ReturnSeparation, analysis.Low, 12}})
}

func TestUnrecoverableAndMissingSyntax(t *testing.T) {
	for _, source := range []string{"def broken(:\n", "def broken():\n    value = (1 + 2\n"} {
		result := Analyze(context.Background(), []byte(source), "reported/source.py", explicitSettings())
		if result.Coverage.Status != analysis.Skipped || len(result.Errors) == 0 || result.Errors[0].Path != "reported/source.py" || result.Errors[0].Line < 1 {
			t.Errorf("syntax failure coverage/errors = %#v / %#v", result.Coverage, result.Errors)
		}
		assertFindings(t, result, nil)
	}
}

func TestRuleSettingsAndLimits(t *testing.T) {
	settings := explicitSettings()
	settings.StatementRun.Enabled = false
	settings.ControlCuddle.Enabled = false
	settings.ReturnSeparation.Level = analysis.High
	result := Analyze(context.Background(), fixture(t, "dense"), "reported/source.py", settings)
	assertChecked(t, result)
	assertFindings(t, result, []observedFinding{{analysis.ReturnSeparation, analysis.High, 10}})

	settings.StatementRun.Enabled = true
	settings.StatementRun.Level = analysis.Low
	settings.MaxStatements = 8
	settings.ReturnBlockLimit = 9
	result = Analyze(context.Background(), fixture(t, "dense"), "reported/source.py", settings)
	assertFindings(t, result, []observedFinding{{analysis.MaxStatementRun, analysis.Low, 2}})

	settings = explicitSettings()
	settings.StatementRun.Enabled = false
	settings.ReturnSeparation.Enabled = false
	settings.ControlCuddle.Level = analysis.High
	result = Analyze(context.Background(), []byte("def f():\n    prepare()\n    if ready:\n        pass\n"), "reported/source.py", settings)
	assertFindings(t, result, []observedFinding{{analysis.ControlCuddle, analysis.High, 3}})
}

func TestCancelledContextIsNotCleanCoverage(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result := Analyze(ctx, fixture(t, "dense"), "reported/source.py", explicitSettings())
	if result.Coverage.Status != analysis.Skipped || len(result.Errors) != 1 || !strings.Contains(result.Errors[0].Message, context.Canceled.Error()) {
		t.Errorf("cancelled coverage/errors = %#v / %#v", result.Coverage, result.Errors)
	}
	assertFindings(t, result, nil)
}

func TestControlPairOuterBoundary(t *testing.T) {
	for _, tc := range []struct {
		name   string
		source string
		line   int
	}{
		{"crowded pair", "def f():\n    prepare()\n    ready = check()\n    if ready:\n        return ready\n", 3},
		{"separated pair", "def f():\n    prepare()\n\n    ready = check()\n    if ready:\n        return ready\n", 0},
		{"internal blank does not clear crowding", "def f():\n    prepare()\n    ready = check()\n\n    if ready:\n        return ready\n", 3},
		{"isolated internal blank stays accepted", "def f():\n    ready = check()\n\n    if ready:\n        return ready\n", 0},
		{"both boundaries separated", "def f():\n    prepare()\n\n    ready = check()\n\n    if ready:\n        return ready\n", 0},
		{"unrelated separated control", "def f():\n    prepare()\n    ready = check()\n\n    if other:\n        return ready\n", 0},
		{"attached comments stay with their statements", "def f():\n    prepare()\n    # Assignment comment.\n    ready = check()\n    # Check comment.\n    if ready:\n        return ready\n", 4},
		{"loop pair", "def f():\n    prepare()\n    items = load()\n    for item in items:\n        use(item)\n", 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			config := explicitSettings()
			config.StatementRun.Enabled = false
			config.ReturnSeparation.Enabled = false
			config.ControlCuddle.Level = analysis.High
			result := Analyze(context.Background(), []byte(tc.source), "reported/source.py", config)
			assertChecked(t, result)
			var want []observedFinding
			if tc.line != 0 {
				want = []observedFinding{{analysis.ControlCuddle, analysis.High, tc.line}}
			}
			assertFindings(t, result, want)
		})
	}
}

func TestCrowdedUnknownControlTargetRemainsPartial(t *testing.T) {
	config := explicitSettings()
	config.StatementRun.Enabled = false
	config.ReturnSeparation.Enabled = false
	source := []byte("def f():\n    prepare()\n    state.ready = check()\n\n    if state.ready:\n        return state\n")
	result := Analyze(context.Background(), source, "reported/source.py", config)
	if result.Coverage.Status != analysis.Partial || len(result.Errors) != 0 {
		t.Fatalf("uncertain pair coverage=%+v errors=%+v", result.Coverage, result.Errors)
	}
	assertFindings(t, result, nil)
	source = []byte("def f():\n    prepare()\n\n    state.ready = check()\n\n    if state.ready:\n        return state\n")
	result = Analyze(context.Background(), source, "reported/source.py", config)
	assertChecked(t, result)
	assertFindings(t, result, nil)
}
