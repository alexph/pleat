package python

import (
	"context"
	"fmt"

	sitter "github.com/tree-sitter/go-tree-sitter"
	python "github.com/tree-sitter/tree-sitter-python/bindings/go"

	"pleat/internal/analysis"
)

// Analyze checks Python layout without executing or changing the source.
func Analyze(ctx context.Context, source []byte, path string, settings analysis.Settings) analysis.FileResult {
	a := analyzer{
		ctx: ctx, source: source, settings: settings,
		result: analysis.FileResult{Coverage: analysis.Coverage{Path: path, Status: analysis.Checked}},
	}
	if ctx.Err() != nil {
		return a.cancelled()
	}
	parser := sitter.NewParser()
	defer parser.Close()
	if err := parser.SetLanguage(sitter.NewLanguage(python.Language())); err != nil {
		return a.failed(0, fmt.Sprintf("Cannot initialize Python parser: %v", err))
	}
	tree := parser.ParseWithOptions(func(offset int, _ sitter.Point) []byte {
		if offset >= len(source) {
			return nil
		}
		return source[offset:]
	}, nil, &sitter.ParseOptions{ProgressCallback: func(sitter.ParseState) bool { return ctx.Err() != nil }})
	if tree != nil {
		defer tree.Close()
	}
	if ctx.Err() != nil {
		return a.cancelled()
	}
	if tree == nil {
		return a.failed(0, "Python parser did not produce a syntax tree")
	}
	root := tree.RootNode()
	var bad map[uintptr]bool
	if root.HasError() {
		bad = make(map[uintptr]bool)
		a.inspectErrors(root, bad)
	}
	if ctx.Err() != nil {
		return a.cancelled()
	}
	if !bad[root.Id()] && !root.HasError() {
		a.walkRegion(root)
	} else {
		// Only complete top-level definitions establish independent scope after
		// recovery. Never salvage descendants of a damaged definition or block.
		safe := false
		for i, count := uint(0), root.NamedChildCount(); i < count && ctx.Err() == nil; i++ {
			node := root.NamedChild(i)
			switch node.Kind() {
			case "function_definition", "class_definition", "decorated_definition":
				if !bad[node.Id()] && !node.HasError() {
					safe = true
					a.walkRegion(node)
				}
			}
		}
		if len(a.result.Errors) == 0 {
			a.result.Errors = append(a.result.Errors, analysis.RunError{Path: path, Line: 1, Message: "Python syntax errors prevent complete layout analysis"})
		}
		status := analysis.Skipped
		if safe {
			status = analysis.Partial
		}
		a.result.Coverage = analysis.Coverage{Path: path, Status: status, Reason: "Python syntax errors; damaged enclosing regions were not analyzed", Line: a.result.Errors[0].Line}
	}
	if ctx.Err() != nil {
		return a.cancelled()
	}
	return a.result
}

type analyzer struct {
	ctx      context.Context
	source   []byte
	settings analysis.Settings
	result   analysis.FileResult
}

func (a *analyzer) failed(line int, message string) analysis.FileResult {
	a.result.Findings = nil
	a.result.Coverage.Status = analysis.Skipped
	a.result.Coverage.Reason = message
	a.result.Coverage.Line = line
	a.result.Errors = append(a.result.Errors, analysis.RunError{Path: a.result.Coverage.Path, Line: line, Message: message})
	return a.result
}

func (a *analyzer) cancelled() analysis.FileResult {
	return a.failed(0, fmt.Sprintf("Python analysis cancelled: %v", a.ctx.Err()))
}

func (a *analyzer) partial(line int, reason string) {
	if a.result.Coverage.Status == analysis.Checked {
		a.result.Coverage.Status = analysis.Partial
		a.result.Coverage.Line = line
		a.result.Coverage.Reason = reason
	}
}

func (a *analyzer) inspectErrors(root *sitter.Node, bad map[uintptr]bool) {
	cursor := root.Walk()
	defer cursor.Close()
	for a.ctx.Err() == nil {
		node := cursor.Node()
		if node.IsError() || node.IsMissing() {
			line := int(node.StartPosition().Row) + 1
			if len(a.result.Errors) == 0 {
				message := "Python syntax error"
				if node.IsMissing() {
					message = "Python syntax is missing " + node.Kind()
				}
				a.result.Errors = append(a.result.Errors, analysis.RunError{Path: a.result.Coverage.Path, Line: line, Message: message})
			}
			for parent := node; parent != nil && !bad[parent.Id()]; parent = parent.Parent() {
				bad[parent.Id()] = true
			}
		}
		if cursor.GotoFirstChild() {
			continue
		}
		for !cursor.GotoNextSibling() {
			if !cursor.GotoParent() {
				return
			}
		}
	}
}

func (a *analyzer) walkRegion(root *sitter.Node) {
	cursor := root.Walk()
	defer cursor.Close()
	for a.ctx.Err() == nil {
		node := cursor.Node()
		if node.Kind() == "module" || node.Kind() == "block" {
			parent := node.Parent()
			// A match suite contains connected case clauses, not sibling statements.
			if parent == nil || parent.Kind() != "match_statement" {
				a.checkBlock(node)
			}
		}
		if cursor.GotoFirstChild() {
			continue
		}
		for !cursor.GotoNextSibling() {
			if !cursor.GotoParent() {
				return
			}
		}
	}
}
