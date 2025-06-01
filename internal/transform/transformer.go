package transform

import (
	"encoding/json"
	"fmt"

	"github.com/tomdoesdev/brace/internal/ast"
)

// Transform converts the processed AST to JSON
type Transform struct {
	output map[string]interface{}
}

// New creates a new transform instance
func New() *Transform {
	return &Transform{
		output: make(map[string]interface{}),
	}
}

// Transform converts the AST to JSON and returns it as a string
func (t *Transform) Transform(program *ast.Program) (string, error) {
	// Process all statements
	for _, stmt := range program.Statements {
		err := t.processStatement(stmt)
		if err != nil {
			return "", err
		}
	}

	// Convert to JSON
	jsonBytes, err := json.MarshalIndent(t.output, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshaling to JSON: %v", err)
	}

	return string(jsonBytes), nil
}

// processStatement processes a single statement
func (t *Transform) processStatement(stmt ast.Statement) error {
	switch s := stmt.(type) {
	case *ast.AssignmentStatement:
		return t.processAssignment(s)
	case *ast.TableStatement:
		return t.processTable(s)
	case *ast.DirectiveStatement:
		// Directives don't contribute to output (they're processed during analysis)
		return nil
	default:
		return fmt.Errorf("unknown statement type: %T", stmt)
	}
}

// processAssignment processes a key = value assignment
func (t *Transform) processAssignment(stmt *ast.AssignmentStatement) error {
	value, err := t.evaluateExpression(stmt.Value)
	if err != nil {
		return err
	}

	t.output[stmt.Name.Value] = value
	return nil
}

// processTable processes a table statement
func (t *Transform) processTable(stmt *ast.TableStatement) error {
	// Navigate to the correct nested position
	current := t.output

	// Create nested structure for table path
	for i, pathSegment := range stmt.Path {
		if i == len(stmt.Path)-1 {
			// Last segment - this is where we put the table content
			tableContent, err := t.evaluateExpression(stmt.Body)
			if err != nil {
				return err
			}
			current[pathSegment] = tableContent
		} else {
			// Intermediate segment - ensure nested map exists
			if current[pathSegment] == nil {
				current[pathSegment] = make(map[string]interface{})
			}
			current = current[pathSegment].(map[string]interface{})
		}
	}

	return nil
}

// evaluateExpression converts AST expressions to Go values
func (t *Transform) evaluateExpression(expr ast.Expression) (interface{}, error) {
	switch e := expr.(type) {
	case *ast.StringLiteral:
		return e.Value, nil
	case *ast.NumberLiteral:
		return e.Value, nil
	case *ast.BooleanLiteral:
		return e.Value, nil
	case *ast.NullLiteral:
		return nil, nil
	case *ast.ArrayLiteral:
		return t.evaluateArray(e)
	case *ast.ObjectLiteral:
		return t.evaluateObject(e)
	case *ast.Reference:
		// Use the resolved value from the analyzer
		if e.ResolvedValue != nil {
			return e.ResolvedValue, nil
		}
		return nil, fmt.Errorf("unresolved reference: %s", e.Name)
	case *ast.EnvDirective:
		// Use the resolved value from the analyzer
		if e.ResolvedValue != nil {
			return e.ResolvedValue, nil
		}
		return nil, fmt.Errorf("unresolved environment directive: @env(\"%s\")", e.VarName)
	case *ast.Identifier:
		return e.Value, nil
	default:
		return nil, fmt.Errorf("cannot evaluate expression type: %T", expr)
	}
}

// evaluateArray converts array literals to Go slices
func (t *Transform) evaluateArray(arr *ast.ArrayLiteral) (interface{}, error) {
	elements := make([]interface{}, 0, len(arr.Elements))

	for _, element := range arr.Elements {
		value, err := t.evaluateExpression(element)
		if err != nil {
			return nil, err
		}
		elements = append(elements, value)
	}

	return elements, nil
}

// evaluateObject converts object literals to Go maps
func (t *Transform) evaluateObject(obj *ast.ObjectLiteral) (interface{}, error) {
	result := make(map[string]interface{})

	for key, value := range obj.Pairs {
		keyStr, err := t.evaluateExpression(key)
		if err != nil {
			return nil, err
		}

		keyString, ok := keyStr.(string)
		if !ok {
			return nil, fmt.Errorf("object keys must be strings, got %T", keyStr)
		}

		valueResult, err := t.evaluateExpression(value)
		if err != nil {
			return nil, err
		}

		result[keyString] = valueResult
	}

	return result, nil
}
