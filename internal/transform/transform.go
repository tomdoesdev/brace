package transform

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tomdoesdev/brace/internal/ast"
	"gopkg.in/yaml.v3"
)

// OutputFormat represents the supported output formats
type OutputFormat string

const (
	FormatJSON OutputFormat = "json"
	FormatYAML OutputFormat = "yaml"
)

// Transform converts the processed AST to the specified format
type Transform struct {
	output map[string]interface{}
	format OutputFormat
}

// New creates a new transform instance with JSON as default format
func New() *Transform {
	return &Transform{
		output: make(map[string]interface{}),
		format: FormatJSON,
	}
}

// NewWithFormat creates a new transform instance with specified format
func NewWithFormat(format OutputFormat) *Transform {
	return &Transform{
		output: make(map[string]interface{}),
		format: format,
	}
}

// SetFormat sets the output format
func (t *Transform) SetFormat(format OutputFormat) {
	t.format = format
}

// Transform converts the AST to the specified format and returns it as a string
func (t *Transform) Transform(program *ast.Program) (string, error) {
	// Process all statements
	for _, stmt := range program.Statements {
		err := t.processStatement(stmt)
		if err != nil {
			return "", err
		}
	}

	// Convert to the specified format
	switch t.format {
	case FormatJSON:
		return t.toJSON()
	case FormatYAML:
		return t.toYAML()
	default:
		return "", fmt.Errorf("unsupported output format: %s", t.format)
	}
}

// toJSON converts the output to JSON format
func (t *Transform) toJSON() (string, error) {
	jsonBytes, err := json.MarshalIndent(t.output, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshaling to JSON: %v", err)
	}
	return string(jsonBytes), nil
}

// toYAML converts the output to YAML format
func (t *Transform) toYAML() (string, error) {
	yamlBytes, err := yaml.Marshal(t.output)
	if err != nil {
		return "", fmt.Errorf("error marshaling to YAML: %v", err)
	}
	return string(yamlBytes), nil
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
		// Construct full reference name for error
		fullName := e.Name
		if e.Namespace != "" {
			fullName = e.Namespace + "." + e.Name
		}
		return nil, fmt.Errorf("unresolved reference: %s", fullName)
	case *ast.EnvDirective:
		// Use the resolved value from the analyzer
		if e.ResolvedValue != nil {
			return e.ResolvedValue, nil
		}
		return nil, fmt.Errorf("unresolved environment directive: @env(\"%s\")", e.VarName)
	case *ast.Identifier:
		return e.Value, nil
	case *ast.TemplateStringLiteral:
		return t.evaluateTemplateString(e)
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

// evaluateTemplateString processes template string interpolation
func (t *Transform) evaluateTemplateString(template *ast.TemplateStringLiteral) (interface{}, error) {
	var result strings.Builder

	for _, part := range template.Parts {
		if part.IsLiteral {
			result.WriteString(part.Content)
		} else {
			// Evaluate the interpolated expression
			value, err := t.evaluateExpression(part.Expr)
			if err != nil {
				return nil, fmt.Errorf("error in template interpolation: %v", err)
			}

			// Convert to string
			result.WriteString(fmt.Sprintf("%v", value))
		}
	}

	return result.String(), nil
}
