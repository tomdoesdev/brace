package errors

import (
	"fmt"
	"strings"
)

// CompilerError represents a compilation error with detailed location info
type CompilerError struct {
	Message  string
	Line     int
	Column   int
	Source   string
	Filename string
}

// ErrorReporter handles error formatting and display
type ErrorReporter struct {
	source   string
	filename string
	lines    []string
}

// NewErrorReporter creates a new error reporter
func NewErrorReporter(source, filename string) *ErrorReporter {
	return &ErrorReporter{
		source:   source,
		filename: filename,
		lines:    strings.Split(source, "\n"),
	}
}

// ReportError formats and returns a Rust-style error message
func (er *ErrorReporter) ReportError(message string, line, column int) string {
	if line < 1 || line > len(er.lines) {
		return fmt.Sprintf("Error: %s (invalid line number)", message)
	}

	// Get the problematic line (convert to 0-based indexing)
	problemLine := er.lines[line-1]

	var result strings.Builder

	// Error header
	result.WriteString(fmt.Sprintf("error: %s\n", message))
	result.WriteString(fmt.Sprintf("  --> %s:%d:%d\n", er.filename, line, column))
	result.WriteString("   |\n")

	// Line number with padding
	lineNumStr := fmt.Sprintf("%d", line)
	padding := strings.Repeat(" ", len(lineNumStr))

	// Show the line
	result.WriteString(fmt.Sprintf("%s | %s\n", lineNumStr, problemLine))

	// Show the pointer
	pointer := strings.Repeat(" ", column-1) + "^"
	if column <= len(problemLine) {
		// Try to underline the problematic token
		endCol := column
		if endCol < len(problemLine) {
			// Find the end of the current token
			for endCol < len(problemLine) && !isWhitespace(problemLine[endCol]) {
				endCol++
			}
		}
		if endCol > column {
			pointer = strings.Repeat(" ", column-1) + strings.Repeat("^", endCol-column+1)
		}
	}
	result.WriteString(fmt.Sprintf("%s | %s\n", padding, pointer))
	result.WriteString("   |\n")

	return result.String()
}

// ReportMultipleErrors formats multiple errors
func (er *ErrorReporter) ReportMultipleErrors(errors []CompilerError) string {
	var result strings.Builder

	for i, err := range errors {
		if i > 0 {
			result.WriteString("\n")
		}
		result.WriteString(er.ReportError(err.Message, err.Line, err.Column))
	}

	if len(errors) > 1 {
		result.WriteString(fmt.Sprintf("\nFound %d errors\n", len(errors)))
	}

	return result.String()
}

func isWhitespace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r'
}
