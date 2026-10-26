package app

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"pleat/internal/analysis"
)

func gitTest(t *testing.T, dir string, args ...string) {
	t.Helper()
	base := []string{"-C", dir, "-c", "core.hooksPath=/dev/null", "-c", "core.autocrlf=false", "-c", "user.name=Pleat Eval", "-c", "user.email=pleat@example.invalid"}
	output, err := exec.Command("git", append(base, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
}

func TestDiffUsesFinalWorkingTreeAndReportsSkippedPaths(t *testing.T) {
	dir := t.TempDir()
	gitTest(t, dir, "init", "--quiet")
	for _, name := range []string{"net.go", "staged.go", "unstaged.py", "deleted.ts", "rename.go", "notes.txt"} {
		writeInput(t, dir, name, "baseline for "+name+"\n")
	}
	writeInput(t, dir, ".gitignore", "ignored.py\n")
	gitTest(t, dir, "add", ".")
	gitTest(t, dir, "commit", "--quiet", "-m", "baseline")
	writeInput(t, dir, "net.go", "staged change\n")
	writeInput(t, dir, "staged.go", "staged change\n")
	gitTest(t, dir, "add", "net.go", "staged.go")
	writeInput(t, dir, "net.go", "baseline for net.go\n")
	writeInput(t, dir, "staged.go", "final working tree\n")
	writeInput(t, dir, "unstaged.py", "unstaged change\n")
	writeInput(t, dir, "notes.txt", "changed outside scope\n")
	writeInput(t, dir, "untracked.py", "untracked\n")
	writeInput(t, dir, "ignored.py", "ignored\n")
	writeInput(t, dir, "newstaged.ts", "new staged\n")
	gitTest(t, dir, "add", "newstaged.ts")
	gitTest(t, dir, "mv", "rename.go", "renamed name.go")
	if err := os.Remove(filepath.Join(dir, "deleted.ts")); err != nil {
		t.Fatal(err)
	}
	subdir := filepath.Join(dir, "subdir")
	if err := os.Mkdir(subdir, 0700); err != nil {
		t.Fatal(err)
	}
	selected, err := selectFiles(context.Background(), subdir, options{diff: "HEAD"})
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, file := range selected.files {
		paths = append(paths, file.path)
	}
	want := []string{"newstaged.ts", "renamed name.go", "staged.go", "unstaged.py"}
	if !reflect.DeepEqual(paths, want) {
		t.Fatalf("paths=%q, want %q", paths, want)
	}
	for _, file := range selected.files {
		if file.path == "staged.go" {
			content, err := os.ReadFile(file.absolute)
			if err != nil {
				t.Fatal(err)
			}
			if string(content) != "final working tree\n" {
				t.Fatalf("selected index contents rather than working tree: %s", content)
			}
		}
	}
	if len(selected.coverage) != 2 {
		t.Fatalf("coverage=%+v", selected.coverage)
	}
	for _, item := range selected.coverage {
		if item.Status != analysis.Skipped || item.Reason == "" || (item.Path != "deleted.ts" && item.Path != "notes.txt") {
			t.Fatalf("unexpected coverage %+v", item)
		}
	}
	explicit, err := selectFiles(context.Background(), dir, options{files: []string{"ignored.py", "./ignored.py"}})
	if err != nil || len(explicit.files) != 1 || explicit.files[0].path != "ignored.py" {
		t.Fatalf("explicit ignored selection=%+v err=%v", explicit, err)
	}
}

func TestGitPathsWithNewlinesRemainOnePath(t *testing.T) {
	dir := t.TempDir()
	gitTest(t, dir, "init", "--quiet")
	name := "line\nbreak.py"
	writeInput(t, dir, name, "baseline\n")
	gitTest(t, dir, "add", ".")
	gitTest(t, dir, "commit", "--quiet", "-m", "baseline")
	writeInput(t, dir, name, "changed\n")
	selected, err := selectFiles(context.Background(), dir, options{diff: "HEAD"})
	if err != nil || len(selected.files) != 1 || selected.files[0].path != name {
		t.Fatalf("selection=%+v err=%v", selected, err)
	}
}

func TestEmptyAndInvalidDiffAreDifferent(t *testing.T) {
	dir := t.TempDir()
	gitTest(t, dir, "init", "--quiet")
	if _, err := selectFiles(context.Background(), dir, options{diff: "HEAD"}); err == nil {
		t.Fatal("unborn HEAD was accepted")
	}
	writeInput(t, dir, "source.go", "package source\n")
	gitTest(t, dir, "add", ".")
	gitTest(t, dir, "commit", "--quiet", "-m", "baseline")
	selected, err := selectFiles(context.Background(), dir, options{diff: "HEAD"})
	if err != nil || len(selected.files) != 0 || len(selected.coverage) != 0 {
		t.Fatalf("empty selection=%+v err=%v", selected, err)
	}
	for _, base := range []string{"not-a-reference", "--help"} {
		if _, err := selectFiles(context.Background(), dir, options{diff: base}); err == nil {
			t.Fatalf("invalid base %q accepted", base)
		}
	}
}

func TestDiffDoesNotExecuteRepositoryHelpers(t *testing.T) {
	dir := t.TempDir()
	gitTest(t, dir, "init", "--quiet")
	writeInput(t, dir, "source.py", "value = 1\n")
	writeInput(t, dir, ".gitattributes", "*.py filter=probe\n")
	for _, name := range []string{"fs-probe", "clean-probe", "process-probe"} {
		path := writeInput(t, dir, name, "#!/bin/sh\nprintf called >> \"$0.called\"\ncat\n")
		if err := os.Chmod(path, 0700); err != nil {
			t.Fatal(err)
		}
	}
	gitTest(t, dir, "add", ".")
	gitTest(t, dir, "commit", "--quiet", "-m", "baseline")
	gitTest(t, dir, "config", "core.fsmonitor", "./fs-probe")
	gitTest(t, dir, "config", "filter.probe.clean", "./clean-probe")
	gitTest(t, dir, "config", "filter.probe.process", "./process-probe")
	gitTest(t, dir, "config", "filter.probe.required", "true")
	writeInput(t, dir, "source.py", "value = 2\n")
	selected, err := selectFiles(context.Background(), dir, options{diff: "HEAD"})
	if err != nil {
		t.Fatal(err)
	}
	if len(selected.files) != 1 || selected.files[0].path != "source.py" {
		t.Fatalf("selection=%+v", selected)
	}
	for _, name := range []string{"fs-probe", "clean-probe", "process-probe"} {
		if _, err := os.Stat(filepath.Join(dir, name+".called")); !os.IsNotExist(err) {
			t.Fatalf("repository helper %s executed, or marker lookup failed: %v", name, err)
		}
	}
}
