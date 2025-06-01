package parser

import (
	"strings"
	"testing"

	"github.com/tomdoesdev/brace/internal/lexer"
	"github.com/tomdoesdev/brace/internal/token"
)

func TestLexerOutput(t *testing.T) {
	source := `#database {
    host = "localhost"
    port = 5432
}`

	l := lexer.New(source)

	for {
		tok := l.NextToken()
		t.Logf("Token: %s, Literal: %q", tok.Type, tok.Literal)

		if tok.Type == token.EOF {
			break
		}
	}
}

func TestSimpleTable(t *testing.T) {
	source := `#database {
    host = "localhost"
}`

	l := lexer.New(source)
	p := New(l, "", "")
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		for _, err := range p.Errors() {
			t.Errorf("Parser error: %s", err)
		}
		t.FailNow()
	}

	if len(program.Statements) != 1 {
		t.Fatalf("Expected 1 statement, got %d", len(program.Statements))
	}

	t.Logf("Parsed successfully: %s", program.String())
}

func TestBraceDirectiveValidation(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid brace directive",
			source:      `@brace "1.0.0"\nname = "test"`,
			expectError: false,
		},
		{
			name:        "missing brace directive",
			source:      `name = "test"`,
			expectError: true,
			errorMsg:    "BRACE file must start with @brace directive",
		},
		{
			name:        "wrong first directive",
			source:      `@const { NAME = "test" }`,
			expectError: true,
			errorMsg:    "first directive must be @brace",
		},
		{
			name:        "empty file",
			source:      ``,
			expectError: true,
			errorMsg:    "empty BRACE file",
		},
		{
			name:        "comments before brace directive allowed",
			source:      `// This is a comment\n@brace "1.0.0"\nname = "test"`,
			expectError: false,
		},
		{
			name:        "brace directive without version",
			source:      `@brace\nname = "test"`,
			expectError: true,
			errorMsg:    "@brace directive requires exactly one version parameter",
		},
		{
			name:        "brace directive with non-string version",
			source:      `@brace 1.0\nname = "test"`,
			expectError: true,
			errorMsg:    "@brace version must be a string literal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.source)
			p := New(l, tt.source, "test.brace")
			program := p.ParseProgram()

			errors := p.Errors()
			if tt.expectError {
				if len(errors) == 0 {
					t.Errorf("expected error but got none")
				} else if !strings.Contains(errors[0], tt.errorMsg) {
					t.Errorf("expected error containing '%s', got '%s'", tt.errorMsg, errors[0])
				}
			} else {
				if len(errors) > 0 {
					t.Errorf("unexpected error: %s", errors[0])
				}
				if len(program.Statements) == 0 {
					t.Errorf("expected statements but got none")
				}
			}
		})
	}
}
