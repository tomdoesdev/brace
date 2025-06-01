package compiler

import (
	"fmt"

	"github.com/tomdoesdev/brace/internal/analyzer"
	"github.com/tomdoesdev/brace/internal/lexer"
	"github.com/tomdoesdev/brace/internal/parser"
	"github.com/tomdoesdev/brace/internal/transform"
)

// Compiler orchestrates the compilation pipeline with enhanced error reporting
type Compiler struct {
	outputFormat transform.OutputFormat
}

// New creates a new compiler instance with JSON as default format
func New() *Compiler {
	return &Compiler{
		outputFormat: transform.FormatJSON,
	}
}

// NewWithFormat creates a new compiler instance with specified output format
func NewWithFormat(format transform.OutputFormat) *Compiler {
	return &Compiler{
		outputFormat: format,
	}
}

// SetOutputFormat sets the output format for the compiler
func (c *Compiler) SetOutputFormat(format transform.OutputFormat) {
	c.outputFormat = format
}

// CompileFile compiles a BRACE file with enhanced error reporting
func (c *Compiler) CompileFile(source, filename string) (string, error) {
	return c.compileWithFilename(source, filename)
}

// Compile takes BRACE source code and returns the specified format (for backward compatibility)
func (c *Compiler) Compile(source string) (string, error) {
	return c.compileWithFilename(source, "<stdin>")
}

// CompileToFormat compiles source and returns output in the specified format
func (c *Compiler) CompileToFormat(source string, format transform.OutputFormat) (string, error) {
	oldFormat := c.outputFormat
	c.outputFormat = format
	result, err := c.compileWithFilename(source, "<stdin>")
	c.outputFormat = oldFormat
	return result, err
}

// CompileFileToFormat compiles a file and returns output in the specified format
func (c *Compiler) CompileFileToFormat(source, filename string, format transform.OutputFormat) (string, error) {
	oldFormat := c.outputFormat
	c.outputFormat = format
	result, err := c.compileWithFilename(source, filename)
	c.outputFormat = oldFormat
	return result, err
}

// compileWithFilename handles compilation with filename for error reporting
func (c *Compiler) compileWithFilename(source, filename string) (string, error) {
	// Phase 1: Lexical Analysis
	l := lexer.New(source)

	// Phase 2: Parsing with enhanced error reporting
	p := parser.New(l, source, filename)
	program := p.ParseProgram()

	// Check for parsing errors with detailed reporting
	if errors := p.Errors(); len(errors) > 0 {
		return "", fmt.Errorf("parsing errors:\n%s", errors[0])
	}

	// Phase 3: Semantic Analysis
	a := analyzer.New()
	err := a.Analyze(program)
	if err != nil {
		return "", fmt.Errorf("analysis error: %v", err)
	}

	// Phase 4: Code Generation with specified format
	t := transform.NewWithFormat(c.outputFormat)
	output, err := t.Transform(program)
	if err != nil {
		return "", fmt.Errorf("generation error: %v", err)
	}

	return output, nil
}
