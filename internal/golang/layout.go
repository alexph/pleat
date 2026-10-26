package golang

import (
	"fmt"
	"go/ast"
	"go/token"

	"pleat/internal/analysis"
)

func (a *analyzer) separated(previous, current ast.Stmt) bool {
	start := a.fset.PositionFor(previous.End(), false).Offset
	end := a.fset.PositionFor(current.Pos(), false).Offset
	gap := a.source[start:end]
	fullLine, blank := false, true
	blockComment, lineComment := false, false
	for i := 0; i < len(gap); i++ {
		c := gap[i]
		if c == '\n' {
			if fullLine && blank && !blockComment && !lineComment {
				return true
			}
			fullLine, blank, lineComment = true, !blockComment, false
			continue
		}
		if blockComment {
			blank = false
			if c == '*' && i+1 < len(gap) && gap[i+1] == '/' {
				blockComment = false
				i++
			}
			continue
		}
		if lineComment {
			continue
		}
		if c == '/' && i+1 < len(gap) {
			if gap[i+1] == '*' {
				blockComment, blank = true, false
				i++
				continue
			}
			if gap[i+1] == '/' {
				lineComment, blank = true, false
				i++
				continue
			}
		}
		if c != ' ' && c != '\t' && c != '\r' {
			blank = false
		}
	}
	return false
}

func executable(stmt ast.Stmt) bool {
	switch stmt := stmt.(type) {
	case *ast.EmptyStmt:
		return false
	case *ast.DeclStmt:
		decl, ok := stmt.Decl.(*ast.GenDecl)
		return !ok || decl.Tok == token.VAR
	default:
		return true
	}
}

func unlabel(stmt ast.Stmt) ast.Stmt {
	for {
		label, ok := stmt.(*ast.LabeledStmt)
		if !ok {
			return stmt
		}
		stmt = label.Stmt
	}
}

func (a *analyzer) checkBlock(statements []ast.Stmt) {
	total := 0
	for _, stmt := range statements {
		if executable(stmt) {
			total++
		}
	}
	var previous, first ast.Stmt
	run := 0
	flushRun := func() {
		if a.settings.StatementRun.Enabled && run > a.settings.MaxStatements {
			a.finding(analysis.MaxStatementRun, a.settings.StatementRun.Level, first.Pos(),
				fmt.Sprintf("This run contains %d statements. The limit is %d. Consider separating related statement groups.", run, a.settings.MaxStatements))
		}
	}
	for _, stmt := range statements {
		if a.ctx.Err() != nil {
			return
		}
		if !executable(stmt) {
			continue
		}
		separated := previous != nil && a.separated(previous, stmt)
		if separated {
			flushRun()
			run = 0
		}
		if previous != nil && !separated {
			if a.settings.ControlCuddle.Enabled && isControl(stmt) {
				allowed, known := false, true
				if run == 1 {
					allowed, known = assignmentMatches(previous, stmt)
				}
				if !known {
					a.partial(stmt.Pos(), "A Go assignment target could not be matched safely to its following control header.")
				} else if !allowed {
					a.finding(analysis.ControlCuddle, a.settings.ControlCuddle.Level, unlabel(stmt).Pos(),
						"Consider separating this control statement and its attached comments from the preceding statement group.")
				}
			}
			if a.settings.ReturnSeparation.Enabled && total > a.settings.ReturnBlockLimit {
				if ret, ok := unlabel(stmt).(*ast.ReturnStmt); ok {
					a.finding(analysis.ReturnSeparation, a.settings.ReturnSeparation.Level, ret.Pos(),
						fmt.Sprintf("Consider a blank line before this return. Its block contains more than %d executable statements.", a.settings.ReturnBlockLimit))
				}
			}
		}
		if run == 0 {
			first = stmt
		}
		run++
		previous = stmt
	}
	flushRun()
}

func isControl(stmt ast.Stmt) bool {
	switch unlabel(stmt).(type) {
	case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt:
		return true
	default:
		return false
	}
}

// The second result is false only when an assignment exists but its targets
// cannot be classified as names without inferring data flow.
func assignmentMatches(previous, control ast.Stmt) (bool, bool) {
	var targets []ast.Expr
	var declared []*ast.Ident
	switch previous := unlabel(previous).(type) {
	case *ast.AssignStmt:
		targets = previous.Lhs
		for _, target := range targets {
			if _, ok := target.(*ast.Ident); !ok {
				return false, false
			}
		}
	case *ast.DeclStmt:
		decl, ok := previous.Decl.(*ast.GenDecl)
		if !ok || decl.Tok != token.VAR || len(decl.Specs) != 1 {
			return false, true
		}
		spec, ok := decl.Specs[0].(*ast.ValueSpec)
		if !ok {
			return false, false
		}
		if len(spec.Values) == 0 {
			return false, true
		}
		declared = spec.Names
	default:
		return false, true
	}
	matchesName := func(name string) bool {
		if name == "_" {
			return false
		}
		for _, target := range targets {
			if target.(*ast.Ident).Name == name {
				return true
			}
		}
		for _, target := range declared {
			if target.Name == name {
				return true
			}
		}
		return false
	}
	matched := false
	inspect := func(node ast.Node) {
		if node == nil || matched {
			return
		}
		ast.Inspect(node, func(node ast.Node) bool {
			if node == nil || matched {
				return false
			}
			// A closure body in the header is a separate executable region.
			if _, ok := node.(*ast.FuncLit); ok {
				return false
			}
			if ident, ok := node.(*ast.Ident); ok && matchesName(ident.Name) {
				matched = true
				return false
			}
			return true
		})
	}
	switch control := unlabel(control).(type) {
	case *ast.IfStmt:
		inspect(control.Init)
		inspect(control.Cond)
	case *ast.ForStmt:
		inspect(control.Init)
		inspect(control.Cond)
		inspect(control.Post)
	case *ast.RangeStmt:
		inspect(control.Key)
		inspect(control.Value)
		inspect(control.X)
	}
	return matched, true
}

func (a *analyzer) finding(rule string, level analysis.Level, pos token.Pos, comment string) {
	a.result.Findings = append(a.result.Findings, analysis.Finding{
		Rule:    rule,
		Level:   level,
		Path:    a.result.Coverage.Path,
		Line:    a.line(pos),
		Comment: comment,
	})
}

func (a *analyzer) partial(pos token.Pos, reason string) {
	if a.result.Coverage.Status == analysis.Checked {
		a.result.Coverage.Status = analysis.Partial
		a.result.Coverage.Reason = reason
		a.result.Coverage.Line = a.line(pos)
	}
}
