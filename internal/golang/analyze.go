package golang

import (
	"context"
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"

	"pleat/internal/analysis"
)

type analyzer struct {
	ctx      context.Context
	fset     *token.FileSet
	source   []byte
	settings analysis.Settings
	result   analysis.FileResult
}

// Analyze checks layout without type checking or executing the supplied source.
func Analyze(ctx context.Context, source []byte, path string, settings analysis.Settings) analysis.FileResult {
	a := analyzer{
		ctx:      ctx,
		fset:     token.NewFileSet(),
		source:   source,
		settings: settings,
		result:   analysis.FileResult{Coverage: analysis.Coverage{Path: path, Status: analysis.Checked}},
	}
	if a.canceled() {
		return a.result
	}

	file, parseErr := parser.ParseFile(a.fset, path, source, parser.ParseComments|parser.AllErrors|parser.SkipObjectResolution)
	if a.canceled() {
		return a.result
	}
	var errors scanner.ErrorList
	if parseErr != nil {
		if list, ok := parseErr.(scanner.ErrorList); ok {
			errors = list
			for _, entry := range errors {
				line := 0
				if file != nil {
					tf := a.fset.File(file.Pos())
					if tf != nil && entry.Pos.Offset >= 0 && entry.Pos.Offset <= tf.Size() {
						line = a.line(tf.Pos(entry.Pos.Offset))
					}
				}
				a.result.Errors = append(a.result.Errors, analysis.RunError{Path: path, Line: line, Message: entry.Msg})
			}
		} else {
			a.result.Errors = append(a.result.Errors, analysis.RunError{Path: path, Message: parseErr.Error()})
		}
	}
	if file == nil {
		a.result.Coverage.Status = analysis.Skipped
		a.result.Coverage.Reason = "Go parsing did not establish a trustworthy syntax tree."
		return a.result
	}

	safeRegions := 0
	for _, decl := range file.Decls {
		if a.canceled() {
			return a.result
		}
		if parseErr != nil && !a.safeDeclaration(decl, errors, len(errors) == 0) {
			continue
		}
		ast.Inspect(decl, func(node ast.Node) bool {
			if node == nil || a.ctx.Err() != nil {
				return false
			}
			switch node := node.(type) {
			case *ast.BlockStmt:
				safeRegions++
				// Switch and select bodies contain clauses, not executable siblings.
				if len(node.List) > 0 {
					switch node.List[0].(type) {
					case *ast.CaseClause, *ast.CommClause:
						return true
					}
				}
				a.checkBlock(node.List)
			case *ast.CaseClause:
				a.checkBlock(node.Body)
			case *ast.CommClause:
				a.checkBlock(node.Body)
			}
			return true
		})
	}
	if a.canceled() {
		return a.result
	}
	if parseErr != nil {
		a.result.Coverage.Status = analysis.Partial
		a.result.Coverage.Reason = "Go syntax errors prevented analysis of affected declarations."
		if len(a.result.Errors) > 0 {
			a.result.Coverage.Line = a.result.Errors[0].Line
		}
		if safeRegions == 0 {
			a.result.Coverage.Status = analysis.Skipped
			a.result.Coverage.Reason = "Go syntax errors left no independently trustworthy executable region."
		}
	}
	return a.result
}

func (a *analyzer) safeDeclaration(decl ast.Decl, errors scanner.ErrorList, unknownError bool) bool {
	if unknownError || !decl.Pos().IsValid() || !decl.End().IsValid() {
		return false
	}
	start := a.fset.PositionFor(decl.Pos(), false).Offset
	end := a.fset.PositionFor(decl.End(), false).Offset
	if end < start || end > len(a.source) {
		return false
	}
	for _, entry := range errors {
		if entry.Pos.Offset >= start && entry.Pos.Offset <= end {
			return false
		}
	}
	safe := true
	ast.Inspect(decl, func(node ast.Node) bool {
		if node == nil || !safe || a.ctx.Err() != nil {
			return false
		}
		switch node := node.(type) {
		case *ast.BadDecl, *ast.BadExpr, *ast.BadStmt:
			safe = false
		case *ast.BlockStmt:
			if !a.sourceToken(node.Lbrace, '{') || !a.sourceToken(node.Rbrace, '}') {
				safe = false
			}
		}
		return safe
	})
	return safe
}

func (a *analyzer) sourceToken(pos token.Pos, want byte) bool {
	if !pos.IsValid() {
		return false
	}
	position := a.fset.PositionFor(pos, false)
	return position.IsValid() && position.Offset < len(a.source) && a.source[position.Offset] == want
}

func (a *analyzer) canceled() bool {
	if err := a.ctx.Err(); err != nil {
		a.result.Findings = nil
		a.result.Coverage.Status = analysis.Skipped
		a.result.Coverage.Reason = "Go analysis was canceled."
		a.result.Coverage.Line = 0
		a.result.Errors = []analysis.RunError{{Path: a.result.Coverage.Path, Message: err.Error()}}
		return true
	}
	return false
}

func (a *analyzer) line(pos token.Pos) int {
	return a.fset.PositionFor(pos, false).Line
}
