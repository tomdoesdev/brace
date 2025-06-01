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
	var tok token.Token

	l.skipWhitespace()

	// Store position info for the token
	startLine := l.line
	startColumn := l.column
	startPosition := l.position

	switch l.ch {
	case '=':
		tok = l.createToken(token.ASSIGN, string(l.ch), startLine, startColumn, startPosition, 1)
	case ':':
		tok = l.createToken(token.COLON, string(l.ch), startLine, startColumn, startPosition, 1)
	case ',':
		tok = l.createToken(token.COMMA, string(l.ch), startLine, startColumn, startPosition, 1)
	case ';':
		tok = l.createToken(token.SEMICOLON, string(l.ch), startLine, startColumn, startPosition, 1)
	case '(':
		tok = l.createToken(token.LPAREN, string(l.ch), startLine, startColumn, startPosition, 1)
	case ')':
		tok = l.createToken(token.RPAREN, string(l.ch), startLine, startColumn, startPosition, 1)
	case '{':
		tok = l.createToken(token.LBRACE, string(l.ch), startLine, startColumn, startPosition, 1)
	case '}':
		tok = l.createToken(token.RBRACE, string(l.ch), startLine, startColumn, startPosition, 1)
	case '[':
		tok = l.createToken(token.LBRACKET, string(l.ch), startLine, startColumn, startPosition, 1)
	case ']':
		tok = l.createToken(token.RBRACKET, string(l.ch), startLine, startColumn, startPosition, 1)
	case '@':
		tok = l.createToken(token.AT, string(l.ch), startLine, startColumn, startPosition, 1)
	case '#':
		tok = l.createToken(token.HASH, string(l.ch), startLine, startColumn, startPosition, 1)
	case '.':
		tok = l.createToken(token.DOT, string(l.ch), startLine, startColumn, startPosition, 1)
	case '"':
		// Handle both regular strings and triple-quoted strings
		if l.peekChar() == '"' && l.peekCharAt(2) == '"' {
			literal := l.readTripleQuotedString()
			length := l.position - startPosition
			tok = l.createToken(token.STRING, literal, startLine, startColumn, startPosition, length)
		} else {
			literal := l.readString()
			length := l.position - startPosition + 1 // +1 for closing quote
			tok = l.createToken(token.STRING, literal, startLine, startColumn, startPosition, length)
		}
	case '\'':
		// Handle single-quoted strings
		literal := l.readSingleQuotedString()
		length := l.position - startPosition + 1
		tok = l.createToken(token.STRING, literal, startLine, startColumn, startPosition, length)
	case '/':
		// Handle comments
		if l.peekChar() == '/' {
			literal := l.readSingleLineComment()
			length := l.position - startPosition
			tok = l.createToken(token.COMMENT, literal, startLine, startColumn, startPosition, length)
		} else if l.peekChar() == '*' {
			literal := l.readMultiLineComment()
			length := l.position - startPosition
			tok = l.createToken(token.COMMENT, literal, startLine, startColumn, startPosition, length)
		} else {
			tok = l.createToken(token.ILLEGAL, string(l.ch), startLine, startColumn, startPosition, 1)
		}
	case 0:
		tok = l.createToken(token.EOF, "", startLine, startColumn, startPosition, 0)
	default:
		if isLetter(l.ch) {
			literal := l.readIdentifier()
			length := len(literal)
			tokenType := lookupIdent(literal)
			tok = l.createToken(tokenType, literal, startLine, startColumn, startPosition, length)
			// Don't call readChar() here because readIdentifier() already advanced
			return tok
		} else if isDigit(l.ch) || (l.ch == '-' && isDigit(l.peekChar())) {
			literal := l.readNumber()
			length := len(literal)
			tok = l.createToken(token.NUMBER, literal, startLine, startColumn, startPosition, length)
			return tok
		} else {
			// Create a more descriptive error message for unknown characters
			charDesc := fmt.Sprintf("'%c' (0x%02x)", l.ch, l.ch)
			if l.ch < 32 || l.ch > 126 {
				charDesc = fmt.Sprintf("\\x%02x", l.ch)
			}
			tok = l.createToken(token.ILLEGAL, fmt.Sprintf("unexpected character %s", charDesc), startLine, startColumn, startPosition, 1)
		}
	}

	l.readChar()
	return tok
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

// Add this method to help debug unknown characters
func (l *Lexer) debugChar() string {
	if l.ch == 0 {
		return "EOF"
	}
	if l.ch < 32 || l.ch > 126 {
		return fmt.Sprintf("\\x%02x", l.ch)
	}
	return string(l.ch)
}
