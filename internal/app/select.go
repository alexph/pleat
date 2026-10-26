package app

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"pleat/internal/analysis"
)

type selectedFile struct {
	absolute string
	path     string
}

type selection struct {
	files    []selectedFile
	coverage []analysis.Coverage
}

func supportedPath(path string) bool {
	switch filepath.Ext(path) {
	case ".go", ".py", ".ts":
		return true
	default:
		return false
	}
}

func selectFiles(ctx context.Context, cwd string, opts options) (selection, error) {
	if opts.diff != "" {
		return selectDiff(ctx, cwd, opts.diff)
	}
	result := selection{}
	seen := make(map[string]bool, len(opts.files))
	for _, name := range opts.files {
		absolute := name
		if !filepath.IsAbs(absolute) {
			absolute = filepath.Join(cwd, absolute)
		}
		absolute = filepath.Clean(absolute)
		if seen[absolute] {
			continue
		}
		seen[absolute] = true
		if !supportedPath(absolute) {
			return result, fmt.Errorf("unsupported explicit file %q (use .go, .py, or .ts)", name)
		}
		path, err := filepath.Rel(cwd, absolute)
		if err != nil {
			return result, fmt.Errorf("report path for %q: %w", name, err)
		}
		result.files = append(result.files, selectedFile{absolute: absolute, path: filepath.ToSlash(path)})
	}
	sort.Slice(result.files, func(i, j int) bool { return result.files[i].path < result.files[j].path })
	return result, nil
}

func gitCommand(ctx context.Context, cwd string, args ...string) *exec.Cmd {
	arguments := append([]string{"-c", "core.fsmonitor=false", "-c", "core.hooksPath=/dev/null", "-C", cwd}, args...)
	return exec.CommandContext(ctx, "git", arguments...)
}

func gitOutput(ctx context.Context, cwd string, args ...string) ([]byte, error) {
	command := gitCommand(ctx, cwd, args...)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	data, err := command.Output()
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return nil, fmt.Errorf("git: %s", message)
	}
	return data, nil
}

func filterOverrides(ctx context.Context, root string) ([]string, error) {
	command := gitCommand(ctx, root, "config", "--null", "--name-only", "--get-regexp", `^filter\..*\.(clean|process)$`)
	data, err := command.Output()
	if err != nil {
		if exit, ok := err.(*exec.ExitError); ok && exit.ExitCode() == 1 && ctx.Err() == nil {
			return nil, nil
		}
		return nil, fmt.Errorf("read Git filter configuration: %w", err)
	}
	drivers := make(map[string]bool)
	for _, key := range bytes.Split(data, []byte{0}) {
		if len(key) != 0 {
			name := string(key)
			drivers[name[:strings.LastIndexByte(name, '.')]] = true
		}
	}
	names := make([]string, 0, len(drivers))
	for name := range drivers {
		names = append(names, name)
	}
	sort.Strings(names)
	var overrides []string
	for _, name := range names {
		overrides = append(overrides, "-c", name+".clean=", "-c", name+".process=", "-c", name+".required=false")
	}
	return overrides, nil
}

func selectDiff(ctx context.Context, cwd, base string) (selection, error) {
	result := selection{}
	rootBytes, err := gitOutput(ctx, cwd, "rev-parse", "--show-toplevel")
	if err != nil {
		return result, err
	}
	root := strings.TrimSuffix(string(rootBytes), "\n")
	commit, err := gitOutput(ctx, root, "rev-parse", "--verify", "--end-of-options", base+"^{commit}")
	if err != nil {
		return result, fmt.Errorf("resolve diff base %q: %w", base, err)
	}
	unmerged, err := gitOutput(ctx, root, "ls-files", "--unmerged", "-z")
	if err != nil {
		return result, err
	}
	if len(unmerged) != 0 {
		return result, fmt.Errorf("resolve merge conflicts before checking a diff")
	}
	overrides, err := filterOverrides(ctx, root)
	if err != nil {
		return result, err
	}
	arguments := append(overrides, "diff", "--name-status", "-z", "--find-renames", "--no-ext-diff", "--no-textconv", "--no-color", strings.TrimSuffix(string(commit), "\n"), "--")
	data, err := gitOutput(ctx, root, arguments...)
	if err != nil {
		return result, err
	}
	fields := bytes.Split(data, []byte{0})
	for i := 0; i < len(fields) && len(fields[i]) != 0; {
		status := string(fields[i])
		i++
		if i >= len(fields) || len(fields[i]) == 0 {
			return result, fmt.Errorf("incomplete Git diff path record")
		}
		path := string(fields[i])
		i++
		if status[0] == 'R' || status[0] == 'C' {
			if i >= len(fields) || len(fields[i]) == 0 {
				return result, fmt.Errorf("incomplete Git rename record")
			}
			path = string(fields[i])
			i++
		}
		path = filepath.ToSlash(path)
		if status[0] == 'D' {
			result.coverage = append(result.coverage, analysis.Coverage{Path: path, Status: analysis.Skipped, Reason: "deleted file"})
			continue
		}
		if !supportedPath(path) {
			result.coverage = append(result.coverage, analysis.Coverage{Path: path, Status: analysis.Skipped, Reason: "extension outside .go, .py, and .ts scope"})
			continue
		}
		result.files = append(result.files, selectedFile{absolute: filepath.Join(root, filepath.FromSlash(path)), path: path})
	}
	sort.Slice(result.files, func(i, j int) bool { return result.files[i].path < result.files[j].path })
	return result, nil
}
