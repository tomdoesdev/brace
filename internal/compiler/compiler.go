package compiler

import (
	"fmt"

	"github.com/tomdoesdev/brace/internal/analyzer"
	"github.com/tomdoesdev/brace/internal/lexer"
	"github.com/tomdoesdev/brace/internal/parser"
	"github.com/tomdoesdev/brace/internal/transform"
)

// Compiler orchestrates the compilation pipeline with enhanced error reporting
type Compiler struct{}

// New creates a new compiler instance
func New() *Compiler {
	return &Compiler{}
}

// CompileFile compiles a BRACE file with enhanced error reporting
func (c *Compiler) CompileFile(source, filename string) (string, error) {
	return c.compileWithFilename(source, filename)
}

// Compile takes BRACE source code and returns JSON (for backward compatibility)
func (c *Compiler) Compile(source string) (string, error) {
	return c.compileWithFilename(source, "<stdin>")
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
		// TODO: Enhance analyzer error reporting too
		return "", fmt.Errorf("analysis error: %v", err)
	}

	// Phase 4: Code Generation
	t := transform.New()
	output, err := t.Transform(program)
	if err != nil {
		return "", fmt.Errorf("generation error: %v", err)
	}

	return output, nil
}
