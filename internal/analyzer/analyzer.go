package analyzer

import (
	"fmt"
	"os"
	"strconv"

	"github.com/tomdoesdev/brace/internal/ast"
)

// Analyzer performs semantic analysis on the AST
// This includes processing directives and resolving constant references
type Analyzer struct {
	constants map[string]map[string]interface{} // namespace -> name -> value
	errors    []string
}

// New creates a new analyzer instance
func New() *Analyzer {
	return &Analyzer{
		constants: make(map[string]map[string]interface{}),
		errors:    []string{},
	}
}

// Analyze processes the AST and resolves all directives and references
func (a *Analyzer) Analyze(program *ast.Program) error {
	// First pass: process all directives to build symbol tables
	for _, stmt := range program.Statements {
		if directive, ok := stmt.(*ast.DirectiveStatement); ok {
			err := a.processDirective(directive)
			if err != nil {
				a.errors = append(a.errors, err.Error())
			}
		}
	}

	// Second pass: resolve all references
	for _, stmt := range program.Statements {
		a.resolveReferences(stmt)
	}

	if len(a.errors) > 0 {
		return fmt.Errorf("analysis errors: %v", a.errors)
	}

	return nil
}

// processDirective handles directive execution
func (a *Analyzer) processDirective(directive *ast.DirectiveStatement) error {
	switch directive.Name {
	case "const":
		return a.processConstDirective(directive)
	case "env":
		// env directives are processed during reference resolution
		return nil
	case "brace":
		// Version directive - validate version
		if len(directive.Parameters) != 1 {
			return fmt.Errorf("@brace directive expects exactly one parameter")
		}
		// For now, just accept any version string
		return nil
	default:
		return fmt.Errorf("unknown directive: %s", directive.Name)
	}
}

// processConstDirective processes @const directives
func (a *Analyzer) processConstDirective(directive *ast.DirectiveStatement) error {
	namespace := "global" // default namespace

	// Check if a namespace was specified
	if len(directive.Parameters) > 0 {
		if str, ok := directive.Parameters[0].(*ast.StringLiteral); ok {
			namespace = str.Value
		}
	}

	// Initialize namespace if it doesn't exist
	if a.constants[namespace] == nil {
		a.constants[namespace] = make(map[string]interface{})
	}

	// Process each constant in the body
	for key, value := range directive.Body.Pairs {
		if ident, ok := key.(*ast.Identifier); ok {
			resolvedValue, err := a.evaluateExpression(value)
			if err != nil {
				return fmt.Errorf("error evaluating constant %s: %v", ident.Value, err)
			}
			a.constants[namespace][ident.Value] = resolvedValue
		}
	}

	return nil
}

// evaluateExpression evaluates an expression to get its actual value
func (a *Analyzer) evaluateExpression(expr ast.Expression) (interface{}, error) {
	switch e := expr.(type) {
	case *ast.StringLiteral:
		return e.Value, nil
	case *ast.NumberLiteral:
		return e.Value, nil
	case *ast.BooleanLiteral:
		return e.Value, nil
	case *ast.NullLiteral:
		return nil, nil
	case *ast.EnvDirective:
		return a.evaluateEnvDirectiveExpression(e)
	case *ast.Reference:
		// Handle references to constants that might contain @env values
		return a.resolveReferenceValue(e)
	default:
		return nil, fmt.Errorf("cannot evaluate expression type: %T", expr)
	}
}

// resolveReferenceValue resolves a reference and returns its value
func (a *Analyzer) resolveReferenceValue(ref *ast.Reference) (interface{}, error) {
	namespace := ref.Namespace
	if namespace == "" {
		namespace = "global"
	}

	if ns, exists := a.constants[namespace]; exists {
		if value, exists := ns[ref.Name]; exists {
			return value, nil
		}
	}

	return nil, fmt.Errorf("undefined reference: %s.%s", namespace, ref.Name)
}

// resolveReferences recursively resolves all references in the AST
func (a *Analyzer) resolveReferences(node ast.Node) {
	switch n := node.(type) {
	case *ast.Program:
		for _, stmt := range n.Statements {
			a.resolveReferences(stmt)
		}
	case *ast.AssignmentStatement:
		a.resolveReferences(n.Value)
	case *ast.TableStatement:
		a.resolveReferences(n.Body)
	case *ast.ObjectLiteral:
		for _, value := range n.Pairs {
			a.resolveReferences(value)
		}
	case *ast.ArrayLiteral:
		for _, element := range n.Elements {
			a.resolveReferences(element)
		}
	case *ast.Reference:
		// This is where we resolve constant references
		a.resolveReference(n)
	case *ast.EnvDirective:
		// Resolve environment directives
		a.resolveEnvDirective(n)
	}
}

// resolveReference resolves a single constant reference
func (a *Analyzer) resolveReference(ref *ast.Reference) {
	namespace := ref.Namespace
	if namespace == "" {
		namespace = "global"
	}

	if ns, exists := a.constants[namespace]; exists {
		if value, exists := ns[ref.Name]; exists {
			// Store the resolved value for later use
			ref.ResolvedValue = value
			return
		}
	}

	a.errors = append(a.errors, fmt.Sprintf("undefined reference: %s.%s", namespace, ref.Name))
}

// resolveEnvDirective resolves @env directives by evaluating them
func (a *Analyzer) resolveEnvDirective(env *ast.EnvDirective) {
	value, err := a.evaluateEnvDirectiveExpression(env)
	if err != nil {
		a.errors = append(a.errors, err.Error())
		return
	}

	// Store the resolved value for later use
	env.ResolvedValue = value
}

// evaluateEnvDirectiveExpression evaluates @env directives
func (a *Analyzer) evaluateEnvDirectiveExpression(env *ast.EnvDirective) (interface{}, error) {
	// Get environment variable
	value := os.Getenv(env.VarName)

	if value == "" {
		// Check if default value was provided
		if env.DefaultValue != nil {
			return a.evaluateExpression(env.DefaultValue)
		} else {
			return nil, fmt.Errorf("environment variable %s not set and no default provided", env.VarName)
		}
	}

	// Try to convert to appropriate type based on the value
	if value == "true" {
		return true, nil
	}
	if value == "false" {
		return false, nil
	}
	if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
		return intVal, nil
	}
	if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
		return floatVal, nil
	}

	// Default to string
	return value, nil
}

// Errors returns the collected errors
func (a *Analyzer) Errors() []string {
	return a.errors
}
