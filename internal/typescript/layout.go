package typescript

import (
	"bytes"
	"fmt"

	sitter "github.com/tree-sitter/go-tree-sitter"

	"pleat/internal/analysis"
)

func (a *analyzer) checkBlock(block *sitter.Node) {
	var statements []*sitter.Node
	for i := range block.NamedChildCount() {
		child := block.NamedChild(i)
		if block.Kind() == "switch_case" && block.FieldNameForNamedChild(uint32(i)) == "value" {
			continue
		}
		executable, known := executableStatement(child)
		if !known {
			a.limitation("Unrecognized TypeScript statement; enclosing block layout was not checked", int(child.StartPosition().Row)+1)
			return
		}
		if executable {
			statements = append(statements, child)
		}
	}
	a.blocks++
	runStart := 0
	previousCrowded := false
	for i, statement := range statements {
		if a.ctx.Err() != nil {
			return
		}
		separated := i == 0 || a.blankBetween(statements[i-1], statement)
		if separated {
			a.checkRun(statements, runStart, i)
			runStart = i
		}
		if i > 0 && a.settings.ControlCuddle.Enabled && control(statement) && (!separated || previousCrowded) {
			previous := statements[i-1]
			names, known := assignmentNames(previous)
			matched := false
			for _, name := range names {
				if a.headerMentions(statement, name) {
					matched = true
					break
				}
			}
			switch {
			case !known:
				a.limitation("Assignment target cannot be classified for layout.control-cuddle; this group boundary was not checked", int(statement.StartPosition().Row)+1)
			case matched && previousCrowded:
				a.finding(analysis.ControlCuddle, a.settings.ControlCuddle.Level, previous,
					"Consider separating this assignment and its following control statement from the preceding group. Put the separator before the assignment and any attached comments, not inside the pair.")
			case !matched && !separated:
				a.finding(analysis.ControlCuddle, a.settings.ControlCuddle.Level, statement,
					"Consider separating this control statement from the preceding statement group. Keep any attached comments with it.")
			}
		}
		if !separated && statement.Kind() == "return_statement" && a.settings.ReturnSeparation.Enabled && len(statements) > a.settings.ReturnBlockLimit {
			a.finding(analysis.ReturnSeparation, a.settings.ReturnSeparation.Level, statement, fmt.Sprintf("Consider separating this return from the preceding group. Keep any attached comments with it. Its block contains more than %d executable statements.", a.settings.ReturnBlockLimit))
		}
		previousCrowded = i > 0 && !separated
	}
	a.checkRun(statements, runStart, len(statements))
}

func (a *analyzer) checkRun(statements []*sitter.Node, start, end int) {
	if a.settings.StatementRun.Enabled && end-start > a.settings.MaxStatements {
		a.finding(analysis.MaxStatementRun, a.settings.StatementRun.Level, statements[start], fmt.Sprintf("This run contains %d statements. The configured limit is %d. Consider separating related statement groups.", end-start, a.settings.MaxStatements))
	}
}

func (a *analyzer) finding(rule string, level analysis.Level, node *sitter.Node, comment string) {
	a.result.Findings = append(a.result.Findings, analysis.Finding{Rule: rule, Level: level, Path: a.path, Line: int(node.StartPosition().Row) + 1, Comment: comment})
}

func executableStatement(node *sitter.Node) (bool, bool) {
	switch node.Kind() {
	case "comment", "hash_bang_line", "empty_statement", "interface_declaration", "type_alias_declaration", "function_signature", "ambient_declaration":
		return false, true
	case "export_statement", "import_statement":
		for i := range node.ChildCount() {
			if kind := node.Child(i).Kind(); kind == "type" || kind == "typeof" {
				return false, true
			}
		}
		if declaration := node.ChildByFieldName("declaration"); declaration != nil {
			return executableStatement(declaration)
		}
		return true, true
	case "break_statement", "continue_statement", "debugger_statement", "do_statement", "expression_statement", "for_in_statement", "for_statement", "if_statement", "labeled_statement", "return_statement", "statement_block", "switch_statement", "throw_statement", "try_statement", "while_statement", "with_statement", "abstract_class_declaration", "class_declaration", "enum_declaration", "function_declaration", "generator_function_declaration", "import_alias", "internal_module", "lexical_declaration", "module", "variable_declaration":
		return true, true
	default:
		return false, false
	}
}

func control(node *sitter.Node) bool {
	switch node.Kind() {
	case "if_statement", "for_statement", "for_in_statement", "while_statement", "do_statement":
		return true
	}
	return false
}

// Only complete physical lines outside comments can separate sibling statements.
// Empty lines inside a multiline statement never reach this gap scanner.
func (a *analyzer) blankBetween(previous, next *sitter.Node) bool {
	gap := a.source[previous.EndByte():next.StartByte()]
	fullLine, blank := previous.EndPosition().Column == 0, true
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

// An empty name list means a known non-assignment. Property and indexed writes
// do not bind a name, so their direct-control exception remains unclassified.
func assignmentNames(statement *sitter.Node) ([]*sitter.Node, bool) {
	if statement.Kind() == "export_statement" {
		if declaration := statement.ChildByFieldName("declaration"); declaration != nil {
			return assignmentNames(declaration)
		}
	}
	var names []*sitter.Node
	switch statement.Kind() {
	case "lexical_declaration", "variable_declaration":
		for i := range statement.NamedChildCount() {
			declaration := statement.NamedChild(i)
			if declaration.Kind() != "variable_declarator" || declaration.ChildByFieldName("value") == nil {
				continue
			}
			var known bool
			names, known = bindingNames(declaration.ChildByFieldName("name"), names)
			if !known {
				return nil, false
			}
		}
		return names, true
	case "expression_statement":
		if statement.NamedChildCount() == 0 {
			return nil, true
		}
		expression := unwrap(statement.NamedChild(0))
		switch expression.Kind() {
		case "assignment_expression", "augmented_assignment_expression":
			return bindingNames(expression.ChildByFieldName("left"), nil)
		case "sequence_expression":
			// Several comma-separated writes are not one assignment statement.
			return nil, false
		}
	}
	return nil, true
}

func unwrap(node *sitter.Node) *sitter.Node {
	for node != nil {
		switch node.Kind() {
		case "parenthesized_expression", "non_null_expression", "as_expression", "satisfies_expression":
			if node.NamedChildCount() > 0 {
				node = node.NamedChild(0)
				continue
			}
		}
		break
	}
	return node
}

func bindingNames(node *sitter.Node, names []*sitter.Node) ([]*sitter.Node, bool) {
	if node == nil {
		return names, false
	}
	node = unwrap(node)
	switch node.Kind() {
	case "identifier", "shorthand_property_identifier_pattern":
		return append(names, node), true
	case "assignment_pattern", "object_assignment_pattern":
		return bindingNames(node.ChildByFieldName("left"), names)
	case "pair_pattern":
		return bindingNames(node.ChildByFieldName("value"), names)
	case "array_pattern", "object_pattern", "rest_pattern":
		for i := range node.NamedChildCount() {
			child := node.NamedChild(i)
			if child.Kind() == "comment" {
				continue
			}
			var known bool
			names, known = bindingNames(child, names)
			if !known {
				return names, false
			}
		}
		return names, true
	}
	return names, false
}

func (a *analyzer) headerMentions(node, name *sitter.Node) bool {
	for i := range node.NamedChildCount() {
		field := node.FieldNameForNamedChild(uint32(i))
		switch field {
		case "condition", "initializer", "increment", "left", "right", "value":
			if a.mentions(node.NamedChild(i), a.source[name.StartByte():name.EndByte()]) {
				return true
			}
		}
	}
	return false
}

func (a *analyzer) mentions(node *sitter.Node, name []byte) bool {
	if node.Kind() == "identifier" || node.Kind() == "shorthand_property_identifier" {
		return bytes.Equal(a.source[node.StartByte():node.EndByte()], name)
	}
	for i := range node.NamedChildCount() {
		if a.mentions(node.NamedChild(i), name) {
			return true
		}
	}
	return false
}
