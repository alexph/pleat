package python

import (
	"bytes"
	"fmt"

	sitter "github.com/tree-sitter/go-tree-sitter"

	"pleat/internal/analysis"
)

func (a *analyzer) checkBlock(block *sitter.Node) {
	parent := block.Parent()
	docstrings := block.Kind() == "module" || parent != nil && (parent.Kind() == "function_definition" || parent.Kind() == "class_definition")
	count := 0
	first := true
	for i, n := uint(0), block.NamedChildCount(); i < n && a.ctx.Err() == nil; i++ {
		node := block.NamedChild(i)
		if node.Kind() == "comment" {
			continue
		}
		if !statementKind(node.Kind()) {
			a.partial(int(node.StartPosition().Row)+1, "Unsupported Python statement classification; enclosing block was not analyzed")
			return
		}
		if !(first && docstrings && a.isDocstring(node)) {
			count++
		}
		first = false
	}
	var previous, runStart *sitter.Node
	previousCrowded := false
	runLength := 0
	first = true
	flushRun := func() {
		if a.settings.StatementRun.Enabled && runLength > a.settings.MaxStatements {
			a.finding(analysis.MaxStatementRun, a.settings.StatementRun.Level, runStart,
				fmt.Sprintf("This run contains %d statements. The configured limit is %d. Consider separating related statement groups.", runLength, a.settings.MaxStatements))
		}
	}
	for i, n := uint(0), block.NamedChildCount(); i < n && a.ctx.Err() == nil; i++ {
		node := block.NamedChild(i)
		if node.Kind() == "comment" {
			continue
		}
		docstring := first && docstrings && a.isDocstring(node)
		first = false
		if docstring {
			continue
		}
		separated := previous == nil || a.blankBetween(previous, node)
		if separated {
			flushRun()
			runStart = node
			runLength = 0
		}
		if previous != nil && a.settings.ControlCuddle.Enabled && isControl(node) && (!separated || previousCrowded) {
			assignment, known, matches := a.assignmentMatches(previous, controlHeader(node))
			matched := assignment && matches
			switch {
			case !known:
				a.partial(int(previous.StartPosition().Row)+1, "Unsupported Python assignment target; control-group check was skipped")
			case matched && previousCrowded:
				a.finding(analysis.ControlCuddle, a.settings.ControlCuddle.Level, previous,
					"Consider separating this assignment and its following control statement from the preceding group. Put the separator before the assignment and any attached comments, not inside the pair.")
			case !matched && !separated:
				a.finding(analysis.ControlCuddle, a.settings.ControlCuddle.Level, node,
					"Consider separating this control statement from the preceding statement group. Keep any attached comments with it.")
			}
		}
		if !separated && a.settings.ReturnSeparation.Enabled && node.Kind() == "return_statement" && count > a.settings.ReturnBlockLimit {
			a.finding(analysis.ReturnSeparation, a.settings.ReturnSeparation.Level, node,
				fmt.Sprintf("Consider separating this return from the preceding group. Keep any attached comments with it. Its block contains more than %d executable statements.", a.settings.ReturnBlockLimit))
		}
		runLength++
		previousCrowded = previous != nil && !separated
		previous = node
	}
	flushRun()
}

func statementKind(kind string) bool {
	switch kind {
	case "future_import_statement", "import_statement", "import_from_statement", "assert_statement", "expression_statement",
		"return_statement", "delete_statement", "raise_statement", "pass_statement", "break_statement", "continue_statement",
		"global_statement", "nonlocal_statement", "type_alias_statement", "if_statement", "for_statement", "while_statement",
		"try_statement", "with_statement", "function_definition", "class_definition", "decorated_definition", "match_statement":
		return true
	default:
		return false
	}
}

func (a *analyzer) finding(rule string, level analysis.Level, node *sitter.Node, comment string) {
	a.result.Findings = append(a.result.Findings, analysis.Finding{
		Rule: rule, Level: level, Path: a.result.Coverage.Path,
		Line: int(node.StartPosition().Row) + 1, Comment: comment,
	})
}

// Only complete physical lines in the source gap can separate siblings.
func (a *analyzer) blankBetween(left, right *sitter.Node) bool {
	gap := a.source[left.EndByte():right.StartByte()]
	completeLine, whitespace := false, true
	for _, c := range gap {
		if c == '\n' {
			if completeLine && whitespace {
				return true
			}
			completeLine, whitespace = true, true
		} else if c != ' ' && c != '\t' && c != '\r' && c != '\f' {
			whitespace = false
		}
	}
	return false
}

func (a *analyzer) isDocstring(node *sitter.Node) bool {
	if node.Kind() == "expression_statement" || node.Kind() == "parenthesized_expression" {
		var expression *sitter.Node
		for i, n := uint(0), node.NamedChildCount(); i < n; i++ {
			child := node.NamedChild(i)
			if child.Kind() == "comment" {
				continue
			}
			if expression != nil {
				return false
			}
			expression = child
		}
		return expression != nil && a.isDocstring(expression)
	}
	if node.Kind() == "concatenated_string" {
		for i, n := uint(0), node.NamedChildCount(); i < n; i++ {
			child := node.NamedChild(i)
			if child.Kind() != "comment" && !a.isDocstring(child) {
				return false
			}
		}
		return true
	}
	if node.Kind() != "string" {
		return false
	}
	for _, c := range a.source[node.StartByte():node.EndByte()] {
		if c == '\'' || c == '"' {
			return true
		}
		if c == 'b' || c == 'B' || c == 'f' || c == 'F' {
			return false
		}
	}
	return false
}

func isControl(node *sitter.Node) bool {
	switch node.Kind() {
	case "if_statement", "while_statement", "for_statement":
		return true
	default:
		return false
	}
}

func controlHeader(node *sitter.Node) *sitter.Node {
	if node.Kind() == "for_statement" {
		return node
	}
	return node.ChildByFieldName("condition")
}

func (a *analyzer) assignmentMatches(statement, header *sitter.Node) (assignment, known, matches bool) {
	if statement.Kind() == "type_alias_statement" {
		return true, false, false
	}
	if statement.Kind() != "expression_statement" || statement.NamedChildCount() != 1 {
		return false, true, false
	}
	node := statement.NamedChild(0)
	if node.Kind() == "named_expression" {
		known, matches = a.targetMatches(node.ChildByFieldName("name"), header)
		return true, known, matches
	}
	if node.Kind() != "assignment" && node.Kind() != "augmented_assignment" {
		return false, true, false
	}
	if node.ChildByFieldName("right") == nil {
		// An annotation without a value does not bind the annotated name.
		return false, true, false
	}
	known = true
	for node != nil && (node.Kind() == "assignment" || node.Kind() == "augmented_assignment") {
		classified, used := a.targetMatches(node.ChildByFieldName("left"), header)
		known = known && classified
		matches = matches || used
		node = node.ChildByFieldName("right")
	}
	return true, known, matches
}

func (a *analyzer) targetMatches(target, header *sitter.Node) (known, matches bool) {
	if target == nil || a.ctx.Err() != nil {
		return false, false
	}
	switch target.Kind() {
	case "identifier":
		return true, a.headerContains(header, a.source[target.StartByte():target.EndByte()])
	case "pattern_list", "tuple_pattern", "list_pattern", "list_splat_pattern", "parenthesized_expression":
		known = true
		for i, n := uint(0), target.NamedChildCount(); i < n; i++ {
			child := target.NamedChild(i)
			if child.Kind() == "comment" {
				continue
			}
			classified, used := a.targetMatches(child, header)
			known = known && classified
			matches = matches || used
		}
		return known, matches
	default:
		// Attribute and subscript writes do not establish a simple assigned
		// name. Preserve that limitation instead of guessing a binding.
		return false, false
	}
}

func (a *analyzer) headerContains(node *sitter.Node, name []byte) bool {
	if node == nil || a.ctx.Err() != nil {
		return false
	}
	if node.Kind() == "for_statement" {
		return a.headerContains(node.ChildByFieldName("left"), name) || a.headerContains(node.ChildByFieldName("right"), name)
	}
	if node.Kind() == "identifier" && bytes.Equal(a.source[node.StartByte():node.EndByte()], name) {
		return true
	}
	for i, n := uint(0), node.NamedChildCount(); i < n; i++ {
		if a.headerContains(node.NamedChild(i), name) {
			return true
		}
	}
	return false
}
