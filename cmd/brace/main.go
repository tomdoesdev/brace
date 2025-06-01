package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/tomdoesdev/brace/internal/compiler"
	"github.com/tomdoesdev/brace/internal/transform"
)

func setupFlags() (*string, *string, *bool, *bool) {
	outputFormat := flag.String("format", "json", "Output format: json or yaml")
	outputFile := flag.String("output", "", "Output file (default: stdout)")
	showHelp := flag.Bool("help", false, "Show help")
	showVersion := flag.Bool("version", false, "Show version")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <file.brace>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Compile BRACE configuration files to JSON or YAML.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s config.brace                    # Output JSON to stdout\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -format=yaml config.brace       # Output YAML to stdout\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -output=config.json config.brace # Output JSON to file\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -format=yaml -output=config.yaml config.brace # Output YAML to file\n", os.Args[0])
	}

	return outputFormat, outputFile, showHelp, showVersion
}

func handleFlags(showHelp, showVersion *bool) string {
	flag.Parse()

	if *showHelp {
		flag.Usage()
		os.Exit(0)
	}

	if *showVersion {
		fmt.Println("brace compiler version 1.0.0")
		os.Exit(0)
	}

	if len(flag.Args()) < 1 {
		fmt.Fprintf(os.Stderr, "Error: No input file specified\n\n")
		flag.Usage()
		os.Exit(1)
	}

	return flag.Args()[0]
}

func determineOutputFormat(outputFormat, outputFile *string) transform.OutputFormat {
	var format transform.OutputFormat
	switch strings.ToLower(*outputFormat) {
	case "json":
		format = transform.FormatJSON
	case "yaml", "yml":
		format = transform.FormatYAML
	default:
		fmt.Fprintf(os.Stderr, "Error: Unsupported output format '%s'. Supported formats: json, yaml\n", *outputFormat)
		os.Exit(1)
	}

	// Auto-detect format from output file extension if not explicitly specified
	if *outputFile != "" && *outputFormat == "json" {
		ext := strings.ToLower(filepath.Ext(*outputFile))
		if ext == ".yaml" || ext == ".yml" {
			format = transform.FormatYAML
		}
	}

	return format
}

func readSourceFile(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", fmt.Errorf("opening file: %v", err)
	}
	defer file.Close()

	source, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("reading file: %v", err)
	}

	return string(source), nil
}

func writeOutput(output, outputFile string) {
	if outputFile != "" {
		err := os.WriteFile(outputFile, []byte(output), 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Output written to %s\n", outputFile)
	} else {
		fmt.Print(output)
	}
}

func main() {
	outputFormat, outputFile, showHelp, showVersion := setupFlags()
	filename := handleFlags(showHelp, showVersion)
	format := determineOutputFormat(outputFormat, outputFile)

	source, err := readSourceFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error %v\n", err)
		os.Exit(1)
	}

	c := compiler.NewWithFormat(format)
	output, err := c.CompileFile(source, filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Compilation error:\n%s\n", err)
		os.Exit(1)
	}

	writeOutput(output, *outputFile)
}
