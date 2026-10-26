package typescript

import (
	"context"
	"fmt"

	sitter "github.com/tree-sitter/go-tree-sitter"
	ts "github.com/tree-sitter/tree-sitter-typescript/bindings/go"

	"pleat/internal/analysis"
)

// Analyze checks plain TypeScript without evaluating or type-checking the source.
func Analyze(ctx context.Context, source []byte, path string, settings analysis.Settings) analysis.FileResult {
	a := analyzer{ctx: ctx, source: source, path: path, settings: settings}
	a.result.Coverage = analysis.Coverage{Path: path, Status: analysis.Checked}
	if err := ctx.Err(); err != nil {
		return a.failed(err.Error(), 0)
	}
	parser := sitter.NewParser()
	defer parser.Close()
	if err := parser.SetLanguage(sitter.NewLanguage(ts.LanguageTypescript())); err != nil {
		return a.failed(fmt.Sprintf("initialize TypeScript parser: %v", err), 0)
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
	if err := ctx.Err(); err != nil {
		return a.failed(err.Error(), 0)
	}
	if tree == nil {
		return a.failed("TypeScript parser returned no syntax tree", 0)
	}
	root := tree.RootNode()
	if line := a.syntaxError(root); line != 0 {
		a.result.Errors = append(a.result.Errors, analysis.RunError{Path: path, Line: line, Message: "TypeScript syntax error or missing token; affected enclosing regions were not checked"})
		a.limitation("TypeScript syntax error; affected enclosing regions were not checked", line)
		// Only intact top-level siblings are independent of a malformed region.
		for i := range root.NamedChildCount() {
			child := root.NamedChild(i)
			if !child.HasError() && !child.IsError() && !child.IsMissing() {
				a.walk(child)
			}
		}
		if a.blocks == 0 {
			a.result.Coverage.Status = analysis.Skipped
		}
	} else {
		a.walk(root)
	}
	if err := ctx.Err(); err != nil {
		return a.failed(err.Error(), 0)
	}
	return a.result
}

type analyzer struct {
	ctx      context.Context
	source   []byte
	path     string
	settings analysis.Settings
	result   analysis.FileResult
	blocks   int
}

func (a *analyzer) failed(message string, line int) analysis.FileResult {
	return analysis.FileResult{
		Coverage: analysis.Coverage{Path: a.path, Status: analysis.Skipped, Reason: message, Line: line},
		Errors:   []analysis.RunError{{Path: a.path, Line: line, Message: message}},
	}
}

func (a *analyzer) limitation(reason string, line int) {
	if a.result.Coverage.Status == analysis.Checked {
		a.result.Coverage.Status = analysis.Partial
		a.result.Coverage.Reason = reason
		a.result.Coverage.Line = line
	}
}

// Missing punctuation is unnamed, so inspect all children rather than named ones.
func (a *analyzer) syntaxError(node *sitter.Node) int {
	if a.ctx.Err() != nil {
		return 0
	}
	if node.IsError() || node.IsMissing() {
		return int(node.StartPosition().Row) + 1
	}
	if !node.HasError() {
		return 0
	}
	for i := range node.ChildCount() {
		if line := a.syntaxError(node.Child(i)); line != 0 {
			return line
		}
	}
	return int(node.StartPosition().Row) + 1
}

func (a *analyzer) walk(node *sitter.Node) {
	if a.ctx.Err() != nil {
		return
	}
	switch node.Kind() {
	case "ambient_declaration", "interface_declaration", "type_alias_declaration", "function_signature":
		return
	case "program", "statement_block", "switch_case", "switch_default":
		a.checkBlock(node)
	}
	for i := range node.NamedChildCount() {
		a.walk(node.NamedChild(i))
	}
}
