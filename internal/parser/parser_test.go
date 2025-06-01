package parser

import (
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
	p := New(l)
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
