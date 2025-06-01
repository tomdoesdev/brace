package token

// TokenType represents the type of a token
type TokenType int

const (
	// Special tokens
	ILLEGAL TokenType = iota
	EOF
	COMMENT

	// Literals
	IDENT  // identifiers like "app_version"
	STRING // "hello" or """multiline"""
	NUMBER // 123, -456, 78.9

	// Keywords
	TRUE
	FALSE
	NULL

	// Operators
	ASSIGN // =
	COLON  // : (for references)

	// Delimiters
	COMMA     // ,
	SEMICOLON // ;

	LPAREN   // (
	RPAREN   // )
	LBRACE   // {
	RBRACE   // }
	LBRACKET // [
	RBRACKET // ]

	// Special symbols
	AT   // @ (directives)
	HASH // # (tables)
	DOT  // . (namespace separator)
)

// Token represents a single token with enhanced position tracking
type Token struct {
	Type     TokenType
	Literal  string
	Line     int
	Column   int
	Position int
	Length   int // Length of the token for better error reporting
}

// String returns a string representation of the token type
func (t TokenType) String() string {
	switch t {
	case ILLEGAL:
		return "ILLEGAL"
	case EOF:
		return "EOF"
	case COMMENT:
		return "COMMENT"
	case IDENT:
		return "IDENT"
	case STRING:
		return "STRING"
	case NUMBER:
		return "NUMBER"
	case TRUE:
		return "TRUE"
	case FALSE:
		return "FALSE"
	case NULL:
		return "NULL"
	case ASSIGN:
		return "="
	case COLON:
		return ":"
	case COMMA:
		return ","
	case SEMICOLON:
		return ";"
	case LPAREN:
		return "("
	case RPAREN:
		return ")"
	case LBRACE:
		return "{"
	case RBRACE:
		return "}"
	case LBRACKET:
		return "["
	case RBRACKET:
		return "]"
	case AT:
		return "@"
	case HASH:
		return "#"
	case DOT:
		return "."
	default:
		return "UNKNOWN"
	}
}
