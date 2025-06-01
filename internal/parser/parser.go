package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/tomdoesdev/brace/internal/ast"
	"github.com/tomdoesdev/brace/internal/errors"
	"github.com/tomdoesdev/brace/internal/lexer"
	"github.com/tomdoesdev/brace/internal/token"
)

// Parser implements a recursive descent parser with enhanced error reporting
type Parser struct {
	l *lexer.Lexer

	curToken  token.Token
	peekToken token.Token

	errorReporter *errors.ErrorReporter
	errors        []errors.CompilerError
}

// New creates a new parser instance
func New(l *lexer.Lexer, source, filename string) *Parser {
	p := &Parser{
		l:             l,
		errorReporter: errors.NewErrorReporter(source, filename),
		errors:        []errors.CompilerError{},
	}

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

// nextToken advances both curToken and peekToken
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// ParseProgram parses the entire BRACE file and returns the AST
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	// Parse until we reach EOF
	for p.curToken.Type != token.EOF {
		// Skip comments - they don't contribute to the AST
		if p.curToken.Type == token.COMMENT {
			p.nextToken()
			continue
		}

		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

// parseStatement determines what type of statement we're parsing
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.AT:
		return p.parseDirectiveStatement()
	case token.HASH:
		return p.parseTableStatement()
	case token.IDENT:
		return p.parseAssignmentStatement()
	default:
		p.addError(fmt.Sprintf("unexpected token: %s", p.curToken.Type))
		return nil
	}
}

// parseDirectiveStatement parses @directive statements (like @const, @brace)
func (p *Parser) parseDirectiveStatement() *ast.DirectiveStatement {
	stmt := &ast.DirectiveStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = p.curToken.Literal

	// Parse parameters based on directive type
	switch stmt.Name {
	case "const":
		return p.parseConstDirective(stmt)
	case "env":
		p.addError("@env directive cannot be used as a statement, only as an expression")
		return nil
	case "brace":
		return p.parseBraceDirective(stmt)
	default:
		p.addError(fmt.Sprintf("unknown directive: %s", stmt.Name))
		return nil
	}
}

// parseConstDirective parses @const directive statements
func (p *Parser) parseConstDirective(stmt *ast.DirectiveStatement) *ast.DirectiveStatement {
	// @const can have optional namespace: @const "namespace" { ... }
	if p.peekToken.Type == token.STRING {
		p.nextToken()
		param := p.parseExpression()
		if param == nil {
			p.addError("failed to parse @const namespace")
			return nil
		}
		stmt.Parameters = append(stmt.Parameters, param)
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	objLiteral := p.parseObjectLiteral()
	if objLiteral == nil {
		p.addError("failed to parse @const body")
		return nil
	}

	if obj, ok := objLiteral.(*ast.ObjectLiteral); ok {
		stmt.Body = obj
	} else {
		p.addError("@const body must be an object")
		return nil
	}

	return stmt
}

// parseBraceDirective parses @brace directive statements
func (p *Parser) parseBraceDirective(stmt *ast.DirectiveStatement) *ast.DirectiveStatement {
	// @brace "version"
	if !p.expectPeek(token.STRING) {
		return nil
	}
	param := p.parseExpression()
	if param == nil {
		p.addError("failed to parse @brace version")
		return nil
	}
	stmt.Parameters = append(stmt.Parameters, param)

	return stmt
}

// parseAssignmentStatement parses key = value assignments
func (p *Parser) parseAssignmentStatement() *ast.AssignmentStatement {
	stmt := &ast.AssignmentStatement{Token: p.curToken}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()
	stmt.Value = p.parseExpression()

	// Optional semicolon
	if p.peekToken.Type == token.SEMICOLON {
		p.nextToken()
	}

	return stmt
}

// parseExpression parses expressions (values)
func (p *Parser) parseExpression() ast.Expression {
	switch p.curToken.Type {
	case token.IDENT:
		return p.parseIdentifier()
	case token.STRING:
		return p.parseStringLiteral()
	case token.NUMBER:
		return p.parseNumberLiteral()
	case token.TRUE, token.FALSE:
		return p.parseBooleanLiteral()
	case token.NULL:
		return p.parseNullLiteral()
	case token.LBRACE:
		return p.parseObjectLiteral()
	case token.LBRACKET:
		return p.parseArrayLiteral()
	case token.COLON:
		return p.parseReference()
	case token.AT:
		return p.parseDirectiveExpression()
	case token.COMMENT:
		// Comments should be skipped at a higher level, but if we encounter one here, skip it
		p.addError("unexpected comment in expression context")
		return nil
	case token.ILLEGAL:
		p.addError(fmt.Sprintf("illegal token: %s", p.curToken.Literal))
		return nil
	default:
		p.addError(fmt.Sprintf("no parse function for %s found", p.curToken.Type))
		return nil
	}
}

// parseDirectiveExpression parses @directive expressions (like @env used in values)
func (p *Parser) parseDirectiveExpression() ast.Expression {
	atToken := p.curToken // save the @ token

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	if p.curToken.Literal == "env" {
		env := &ast.EnvDirective{Token: atToken}

		if !p.expectPeek(token.LPAREN) {
			return nil
		}

		if !p.expectPeek(token.STRING) {
			return nil
		}

		env.VarName = p.curToken.Literal

		// Check for optional default value
		if p.peekToken.Type == token.COMMA {
			p.nextToken() // consume comma
			p.nextToken() // move to default value
			env.DefaultValue = p.parseExpression()
		}

		if !p.expectPeek(token.RPAREN) {
			return nil
		}

		return env
	}

	p.addError(fmt.Sprintf("unknown directive in expression context: %s", p.curToken.Literal))
	return nil
}

// parseTableStatement parses #table statements
func (p *Parser) parseTableStatement() *ast.TableStatement {
	stmt := &ast.TableStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	// Parse table path (table.subtable.etc)
	stmt.Path = []string{p.curToken.Literal}

	for p.peekToken.Type == token.DOT {
		p.nextToken() // consume dot
		if !p.expectPeek(token.IDENT) {
			return nil
		}
		stmt.Path = append(stmt.Path, p.curToken.Literal)
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	// Parse the object body and check if it's valid
	objLiteral := p.parseObjectLiteral()
	if objLiteral == nil {
		p.addError("failed to parse table body")
		return nil
	}

	// Type assert safely
	if obj, ok := objLiteral.(*ast.ObjectLiteral); ok {
		stmt.Body = obj
	} else {
		p.addError("table body must be an object")
		return nil
	}

	return stmt
}

// parseIdentifier parses identifiers
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

// parseStringLiteral parses string literals
func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

// parseNumberLiteral parses number literals and determines if they're int or float
func (p *Parser) parseNumberLiteral() ast.Expression {
	lit := &ast.NumberLiteral{Token: p.curToken}

	// Try to parse as integer first
	if !strings.Contains(p.curToken.Literal, ".") {
		value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
		if err != nil {
			p.addError(fmt.Sprintf("could not parse %q as integer", p.curToken.Literal))
			return nil
		}
		lit.Value = value
	} else {
		// Parse as float
		value, err := strconv.ParseFloat(p.curToken.Literal, 64)
		if err != nil {
			p.addError(fmt.Sprintf("could not parse %q as float", p.curToken.Literal))
			return nil
		}
		lit.Value = value
	}

	return lit
}

// parseBooleanLiteral parses boolean literals
func (p *Parser) parseBooleanLiteral() ast.Expression {
	return &ast.BooleanLiteral{Token: p.curToken, Value: p.curToken.Type == token.TRUE}
}

// parseNullLiteral parses null literals
func (p *Parser) parseNullLiteral() ast.Expression {
	return &ast.NullLiteral{Token: p.curToken}
}

// parseArrayLiteral parses array literals
func (p *Parser) parseArrayLiteral() ast.Expression {
	arr := &ast.ArrayLiteral{Token: p.curToken}
	arr.Elements = p.parseExpressionList(token.RBRACKET)
	return arr
}

// parseObjectLiteral parses object literals with optional commas and comment support
func (p *Parser) parseObjectLiteral() ast.Expression {
	obj := &ast.ObjectLiteral{Token: p.curToken}
	obj.Pairs = make(map[ast.Expression]ast.Expression)

	if p.peekToken.Type == token.RBRACE {
		p.nextToken()
		return obj
	}

	p.nextToken()

	if !p.parseObjectPairs(obj) {
		return nil
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}

	return obj
}

// parseObjectPairs parses all key-value pairs in an object literal
func (p *Parser) parseObjectPairs(obj *ast.ObjectLiteral) bool {
	for {
		p.skipCommentsInObject()

		if p.curToken.Type == token.RBRACE {
			break
		}

		if !p.parseObjectPair(obj) {
			return false
		}

		p.skipTrailingComments()

		if !p.handleObjectSeparators() {
			break
		}
	}
	return true
}

// parseObjectPair parses a single key-value pair in an object literal
func (p *Parser) parseObjectPair(obj *ast.ObjectLiteral) bool {
	key := p.parseExpression()
	if key == nil {
		p.addError("failed to parse object key")
		return false
	}

	if !p.expectPeek(token.ASSIGN) {
		return false
	}

	p.nextToken()
	value := p.parseExpression()
	if value == nil {
		p.addError("failed to parse object value")
		return false
	}

	obj.Pairs[key] = value
	return true
}

// skipCommentsInObject skips any comments inside the object
func (p *Parser) skipCommentsInObject() {
	for p.curToken.Type == token.COMMENT {
		p.nextToken()
		if p.curToken.Type == token.RBRACE {
			break
		}
	}
}

// skipTrailingComments skips any trailing comments after a value
func (p *Parser) skipTrailingComments() {
	for p.peekToken.Type == token.COMMENT {
		p.nextToken()
	}
}

// handleObjectSeparators handles commas and other separators between object pairs
func (p *Parser) handleObjectSeparators() bool {
	switch p.peekToken.Type {
	case token.COMMA:
		p.nextToken() // consume comma
		if p.peekToken.Type == token.RBRACE {
			return false // trailing comma case
		}
		p.nextToken() // move to next key
		return true
	case token.RBRACE:
		return false // end of object
	case token.IDENT, token.COMMENT:
		// No comma, but we have another identifier (another key) or a comment to skip
		p.nextToken() // move to next key or comment (will be handled by skipCommentsInObject)
		return true
	default:
		// Unexpected token
		return false
	}
}

// parseReference parses constant references like :namespace.CONSTANT
func (p *Parser) parseReference() ast.Expression {
	ref := &ast.Reference{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	// Check if this is a namespaced reference
	if p.peekToken.Type == token.DOT {
		ref.Namespace = p.curToken.Literal
		p.nextToken() // consume dot
		if !p.expectPeek(token.IDENT) {
			return nil
		}
		ref.Name = p.curToken.Literal
	} else {
		ref.Name = p.curToken.Literal
	}

	return ref
}

// parseExpressionList parses a comma-separated list of expressions
func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	args := []ast.Expression{}

	if p.peekToken.Type == end {
		p.nextToken()
		return args
	}

	p.nextToken()
	args = append(args, p.parseExpression())

	for p.peekToken.Type == token.COMMA {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression())
	}

	if !p.expectPeek(end) {
		return nil
	}

	return args
}

// expectPeek checks if the next token is of the expected type and advances if so
func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekToken.Type == t {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

// addError adds an error with position information
func (p *Parser) addError(msg string) {
	err := errors.CompilerError{
		Message:  msg,
		Line:     p.curToken.Line,
		Column:   p.curToken.Column,
		Source:   "",
		Filename: "",
	}
	p.errors = append(p.errors, err)
}

// addErrorAtToken adds an error at a specific token's position
func (p *Parser) addErrorAtToken(msg string, tok token.Token) {
	err := errors.CompilerError{
		Message:  msg,
		Line:     tok.Line,
		Column:   tok.Column,
		Source:   "",
		Filename: "",
	}
	p.errors = append(p.errors, err)
}

// peekError creates an error for unexpected peek token
func (p *Parser) peekError(expected token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead", expected, p.peekToken.Type)
	p.addErrorAtToken(msg, p.peekToken)
}

// Errors returns formatted error messages
func (p *Parser) Errors() []string {
	if len(p.errors) == 0 {
		return []string{}
	}

	formatted := p.errorReporter.ReportMultipleErrors(p.errors)
	return []string{formatted}
}

// GetDetailedErrors returns the raw error objects for more detailed handling
func (p *Parser) GetDetailedErrors() []errors.CompilerError {
	return p.errors
}
