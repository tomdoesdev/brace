package ast

import "github.com/tomdoesdev/brace/internal/token"

// Node represents any node in the AST
// All AST nodes implement this interface
type Node interface {
	TokenLiteral() string // Returns the literal value of the token
	String() string       // String representation for debugging
}

// Statement represents a statement node
// Statements don't return values (assignments, directives, tables)
type Statement interface {
	Node
	statementNode()
}

// Expression represents an expression node
// Expressions return values (strings, numbers, objects, etc.)
type Expression interface {
	Node
	expressionNode()
}

// Program is the root node of every AST
// It contains all statements in the BRACE file
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var out string
	for _, s := range p.Statements {
		out += s.String()
	}
	return out
}

// Assignment represents a key = value assignment
type AssignmentStatement struct {
	Token token.Token // the IDENT token
	Name  *Identifier
	Value Expression
}

func (as *AssignmentStatement) statementNode() { /* marker method for Statement interface */ }
func (as *AssignmentStatement) TokenLiteral() string { return as.Token.Literal }
func (as *AssignmentStatement) String() string {
	return as.Name.String() + " = " + as.Value.String()
}

// Directive represents @directive statements
type DirectiveStatement struct {
	Token      token.Token // the @ token
	Name       string
	Parameters []Expression
	Body       *ObjectLiteral
}

func (ds *DirectiveStatement) statementNode() { /* marker method for Statement interface */ }
func (ds *DirectiveStatement) TokenLiteral() string { return ds.Token.Literal }
func (ds *DirectiveStatement) String() string {
	return "@" + ds.Name
}

// EnvDirective represents @env directives used as expressions
type EnvDirective struct {
	Token         token.Token // the @ token
	VarName       string
	DefaultValue  Expression  // optional default value
	ResolvedValue interface{} // resolved value after analysis
}

func (ed *EnvDirective) expressionNode() { /* marker method for Expression interface */ }
func (ed *EnvDirective) TokenLiteral() string { return ed.Token.Literal }
func (ed *EnvDirective) String() string {
	if ed.DefaultValue != nil {
		return "@env(\"" + ed.VarName + "\", " + ed.DefaultValue.String() + ")"
	}
	return "@env(\"" + ed.VarName + "\")"
}

// Table represents #table statements
type TableStatement struct {
	Token token.Token // the # token
	Path  []string    // table.subtable becomes ["table", "subtable"]
	Body  *ObjectLiteral
}

func (ts *TableStatement) statementNode() { /* marker method for Statement interface */ }
func (ts *TableStatement) TokenLiteral() string { return ts.Token.Literal }
func (ts *TableStatement) String() string {
	return "#" + ts.Path[0]
}

// Identifier represents variable names
type Identifier struct {
	Token token.Token
	Value string
}

func (i *Identifier) expressionNode() { /* marker method for Expression interface */ }
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

// StringLiteral represents string values
type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode() { /* marker method for Expression interface */ }
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return "\"" + sl.Value + "\"" }

// NumberLiteral represents numeric values
type NumberLiteral struct {
	Token token.Token
	Value interface{} // int64 or float64
}

func (nl *NumberLiteral) expressionNode() { /* marker method for Expression interface */ }
func (nl *NumberLiteral) TokenLiteral() string { return nl.Token.Literal }
func (nl *NumberLiteral) String() string       { return nl.Token.Literal }

// BooleanLiteral represents true/false values
type BooleanLiteral struct {
	Token token.Token
	Value bool
}

func (bl *BooleanLiteral) expressionNode() { /* marker method for Expression interface */ }
func (bl *BooleanLiteral) TokenLiteral() string { return bl.Token.Literal }
func (bl *BooleanLiteral) String() string       { return bl.Token.Literal }

// NullLiteral represents null values
type NullLiteral struct {
	Token token.Token
}

func (nl *NullLiteral) expressionNode() { /* marker method for Expression interface */ }
func (nl *NullLiteral) TokenLiteral() string { return nl.Token.Literal }
func (nl *NullLiteral) String() string       { return "null" }

// ArrayLiteral represents arrays
type ArrayLiteral struct {
	Token    token.Token // the '[' token
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode() { /* marker method for Expression interface */ }
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }
func (al *ArrayLiteral) String() string       { return "[...]" }

// ObjectLiteral represents objects
type ObjectLiteral struct {
	Token token.Token // the '{' token
	Pairs map[Expression]Expression
}

func (ol *ObjectLiteral) expressionNode() { /* marker method for Expression interface */ }
func (ol *ObjectLiteral) TokenLiteral() string { return ol.Token.Literal }
func (ol *ObjectLiteral) String() string       { return "{...}" }

// Reference represents constant references like :namespace.CONSTANT
type Reference struct {
	Token         token.Token // the ':' token
	Namespace     string      // optional namespace
	Name          string      // constant name
	ResolvedValue interface{} // resolved value after analysis
}

func (r *Reference) expressionNode() { /* marker method for Expression interface */ }
func (r *Reference) TokenLiteral() string { return r.Token.Literal }
func (r *Reference) String() string {
	if r.Namespace != "" {
		return ":" + r.Namespace + "." + r.Name
	}
	return ":" + r.Name
}
