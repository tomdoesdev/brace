package lexer

import (
	"fmt"

	"github.com/tomdoesdev/brace/internal/token"
)

// Lexer performs lexical analysis with enhanced error reporting
type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
	line         int  // current line number for error reporting
	column       int  // current column number for error reporting
}

// New creates a new lexer instance
func New(input string) *Lexer {
	l := &Lexer{
		input:  input,
		line:   1,
		column: 0,
	}
	l.readChar() // Initialize by reading the first character
	return l
}

// readChar reads the next character and advances position
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // ASCII NUL character represents "EOF"
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++

	if l.ch == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}
}

// NextToken scans the input and returns the next token with position info
func (l *Lexer) NextToken() token.Token {
	l.skipWhitespace()

	// Store position info for the token
	startLine := l.line
	startColumn := l.column
	startPosition := l.position

	tok := l.scanToken(startLine, startColumn, startPosition)

	if tok.Type != token.IDENT && tok.Type != token.NUMBER {
		l.readChar()
	}
	return tok
}

// scanToken handles the main token scanning logic
func (l *Lexer) scanToken(startLine, startColumn, startPosition int) token.Token {
	switch l.ch {
	case '=':
		return l.createToken(token.ASSIGN, string(l.ch), startLine, startColumn, startPosition, 1)
	case ':':
		return l.createToken(token.COLON, string(l.ch), startLine, startColumn, startPosition, 1)
	case ',':
		return l.createToken(token.COMMA, string(l.ch), startLine, startColumn, startPosition, 1)
	case ';':
		return l.createToken(token.SEMICOLON, string(l.ch), startLine, startColumn, startPosition, 1)
	case '(':
		return l.createToken(token.LPAREN, string(l.ch), startLine, startColumn, startPosition, 1)
	case ')':
		return l.createToken(token.RPAREN, string(l.ch), startLine, startColumn, startPosition, 1)
	case '{':
		return l.createToken(token.LBRACE, string(l.ch), startLine, startColumn, startPosition, 1)
	case '}':
		return l.createToken(token.RBRACE, string(l.ch), startLine, startColumn, startPosition, 1)
	case '[':
		return l.createToken(token.LBRACKET, string(l.ch), startLine, startColumn, startPosition, 1)
	case ']':
		return l.createToken(token.RBRACKET, string(l.ch), startLine, startColumn, startPosition, 1)
	case '@':
		return l.createToken(token.AT, string(l.ch), startLine, startColumn, startPosition, 1)
	case '#':
		return l.createToken(token.HASH, string(l.ch), startLine, startColumn, startPosition, 1)
	case '.':
		return l.createToken(token.DOT, string(l.ch), startLine, startColumn, startPosition, 1)
	case '"':
		return l.handleStringToken(startLine, startColumn, startPosition)
	case '\'':
		return l.handleSingleQuotedStringToken(startLine, startColumn, startPosition)
	case '`':
		return l.handleTemplateStringToken(startLine, startColumn, startPosition)
	case '/':
		return l.handleSlashToken(startLine, startColumn, startPosition)
	case 0:
		return l.createToken(token.EOF, "", startLine, startColumn, startPosition, 0)
	default:
		return l.handleDefaultToken(startLine, startColumn, startPosition)
	}
}

// handleStringToken handles double-quoted string tokens
func (l *Lexer) handleStringToken(startLine, startColumn, startPosition int) token.Token {
	if l.peekChar() == '"' && l.peekCharAt(2) == '"' {
		literal := l.readTripleQuotedString()
		length := l.position - startPosition
		return l.createToken(token.STRING, literal, startLine, startColumn, startPosition, length)
	}
	literal := l.readString()
	length := l.position - startPosition + 1
	return l.createToken(token.STRING, literal, startLine, startColumn, startPosition, length)
}

// handleSingleQuotedStringToken handles single-quoted string tokens
func (l *Lexer) handleSingleQuotedStringToken(startLine, startColumn, startPosition int) token.Token {
	literal := l.readSingleQuotedString()
	length := l.position - startPosition + 1
	return l.createToken(token.STRING, literal, startLine, startColumn, startPosition, length)
}

// handleTemplateStringToken handles backtick-quoted template string tokens
func (l *Lexer) handleTemplateStringToken(startLine, startColumn, startPosition int) token.Token {
	literal := l.readTemplateString()
	length := l.position - startPosition + 1
	return l.createToken(token.TEMPLATE_STRING, literal, startLine, startColumn, startPosition, length)
}

// handleSlashToken handles slash tokens (comments or illegal)
func (l *Lexer) handleSlashToken(startLine, startColumn, startPosition int) token.Token {
	if l.peekChar() == '/' {
		literal := l.readSingleLineComment()
		length := l.position - startPosition
		return l.createToken(token.COMMENT, literal, startLine, startColumn, startPosition, length)
	}
	if l.peekChar() == '*' {
		literal := l.readMultiLineComment()
		length := l.position - startPosition
		return l.createToken(token.COMMENT, literal, startLine, startColumn, startPosition, length)
	}
	return l.createToken(token.ILLEGAL, string(l.ch), startLine, startColumn, startPosition, 1)
}

// handleDefaultToken handles identifiers, numbers, and illegal characters
func (l *Lexer) handleDefaultToken(startLine, startColumn, startPosition int) token.Token {
	if isLetter(l.ch) {
		literal := l.readIdentifier()
		length := len(literal)
		tokenType := lookupIdent(literal)
		return l.createToken(tokenType, literal, startLine, startColumn, startPosition, length)
	}
	if isDigit(l.ch) || (l.ch == '-' && isDigit(l.peekChar())) {
		literal := l.readNumber()
		length := len(literal)
		return l.createToken(token.NUMBER, literal, startLine, startColumn, startPosition, length)
	}
	charDesc := fmt.Sprintf("'%c' (0x%02x)", l.ch, l.ch)
	if l.ch < 32 || l.ch > 126 {
		charDesc = fmt.Sprintf("\\x%02x", l.ch)
	}
	return l.createToken(token.ILLEGAL, fmt.Sprintf("unexpected character %s", charDesc), startLine, startColumn, startPosition, 1)
}

// createToken is a helper to create tokens with position info
func (l *Lexer) createToken(tokenType token.TokenType, literal string, line, column, position, length int) token.Token {
	return token.Token{
		Type:     tokenType,
		Literal:  literal,
		Line:     line,
		Column:   column,
		Position: position,
		Length:   length,
	}
}

// peekChar returns the next character without advancing position
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// Helper method to peek at character at specific offset
func (l *Lexer) peekCharAt(offset int) byte {
	pos := l.readPosition + offset - 1
	if pos >= len(l.input) {
		return 0
	}
	return l.input[pos]
}

// skipWhitespace skips whitespace characters
// BRACE is whitespace-insensitive, so we can safely ignore all whitespace
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// readIdentifier reads an identifier (variable names, directive names, etc.)
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
	return l.input[position:l.position]
}

// readNumber reads a number (integer or decimal)
func (l *Lexer) readNumber() string {
	position := l.position

	// Handle negative numbers
	if l.ch == '-' {
		l.readChar()
	}

	// Read integer part
	for isDigit(l.ch) {
		l.readChar()
	}

	// Check for decimal point
	if l.ch == '.' && isDigit(l.peekChar()) {
		l.readChar() // consume '.'
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	return l.input[position:l.position]
}

// readString reads a double-quoted string
func (l *Lexer) readString() string {
	position := l.position + 1 // Skip opening quote
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}
	return l.input[position:l.position]
}

// readSingleQuotedString reads a single-quoted string
func (l *Lexer) readSingleQuotedString() string {
	position := l.position + 1 // Skip opening quote
	for {
		l.readChar()
		if l.ch == '\'' || l.ch == 0 {
			break
		}
	}
	return l.input[position:l.position]
}

// readTripleQuotedString reads a triple-quoted multiline string
func (l *Lexer) readTripleQuotedString() string {
	// Skip opening """
	l.readChar() // first "
	l.readChar() // second "
	l.readChar() // third "

	position := l.position

	for {
		if l.ch == 0 {
			break
		}
		if l.ch == '"' && l.peekChar() == '"' && l.peekCharAt(2) == '"' {
			break
		}
		l.readChar()
	}

	result := l.input[position:l.position]

	// Skip closing """
	l.readChar() // first "
	l.readChar() // second "
	// Third " will be consumed by NextToken's readChar()

	return result
}

// readSingleLineComment reads a single line comment
func (l *Lexer) readSingleLineComment() string {
	position := l.position
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
	return l.input[position:l.position]
}

// readMultiLineComment reads a multi-line comment
func (l *Lexer) readMultiLineComment() string {
	position := l.position

	// Skip /*
	l.readChar()
	l.readChar()

	for {
		if l.ch == 0 {
			break
		}
		if l.ch == '*' && l.peekChar() == '/' {
			l.readChar() // consume *
			l.readChar() // consume /
			break
		}
		l.readChar()
	}

	return l.input[position:l.position]
}

// readTemplateString reads a backtick-quoted template string
func (l *Lexer) readTemplateString() string {
	position := l.position + 1 // Skip opening backtick
	for {
		l.readChar()
		if l.ch == '`' || l.ch == 0 {
			break
		}
	}
	return l.input[position:l.position]
}

// isLetter checks if a character is a letter
func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z'
}

// isDigit checks if a character is a digit
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

// lookupIdent checks if an identifier is a keyword
func lookupIdent(ident string) token.TokenType {
	switch ident {
	case "true":
		return token.TRUE
	case "false":
		return token.FALSE
	case "null":
		return token.NULL
	default:
		return token.IDENT
	}
}
