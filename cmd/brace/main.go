package main

import (
	"fmt"
	"io"
	"os"

	"github.com/tomdoesdev/brace/internal/compiler"
)

func main() {
	// Check if a file was provided as argument
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <file.brace>\n", os.Args[0])
		os.Exit(1)
	}

	filename := os.Args[1]

	// Read the BRACE file
	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	// Read file contents
	source, err := io.ReadAll(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Create compiler and compile the source with filename for better errors
	c := compiler.New()
	output, err := c.CompileFile(string(source), filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Compilation error:\n%s\n", err)
		os.Exit(1)
	}

	// Output JSON to stdout (can be piped)
	fmt.Print(output)
}
