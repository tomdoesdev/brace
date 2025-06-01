package compiler

import (
	"os"
	"strings"
	"testing"

	"github.com/tomdoesdev/brace/internal/transform"
)

func TestBasicCompilationJSON(t *testing.T) {
	source := `
@brace "0.0.1"

@const {
    VERSION = "1.0.0"
}

app_version = :VERSION
name = "test app"
port = 8080
enabled = true

#database {
    host = "localhost"
    port = 5432
}
`

	compiler := New()
	output, err := compiler.Compile(source)
	if err != nil {
		t.Fatalf("JSON compilation failed: %v", err)
	}

	// Check that output contains expected JSON structure
	if !strings.Contains(output, `"app_version": "1.0.0"`) {
		t.Errorf("Expected app_version to be resolved to '1.0.0'")
	}

	if !strings.Contains(output, `"database"`) {
		t.Errorf("Expected database table to be present")
	}

	t.Logf("JSON compilation output:\n%s", output)
}

func TestBasicCompilationYAML(t *testing.T) {
	source := `
@brace "0.0.1"

@const {
    VERSION = "1.0.0"
}

app_version = :VERSION
name = "test app"
port = 8080
enabled = true

#database {
    host = "localhost"
    port = 5432
}
`

	compiler := NewWithFormat(transform.FormatYAML)
	output, err := compiler.Compile(source)
	if err != nil {
		t.Fatalf("YAML compilation failed: %v", err)
	}

	// Check that output contains expected YAML structure
	if !strings.Contains(output, "app_version: 1.0.0") {
		t.Errorf("Expected app_version to be resolved to '1.0.0' in YAML")
	}

	if !strings.Contains(output, "database:") {
		t.Errorf("Expected database table to be present in YAML")
	}

	t.Logf("YAML compilation output:\n%s", output)
}

func TestFormatSwitching(t *testing.T) {
	source := `
name = "test"
value = 42
`

	compiler := New()

	// Test JSON output
	jsonOutput, err := compiler.CompileToFormat(source, transform.FormatJSON)
	if err != nil {
		t.Fatalf("JSON compilation failed: %v", err)
	}

	// Test YAML output
	yamlOutput, err := compiler.CompileToFormat(source, transform.FormatYAML)
	if err != nil {
		t.Fatalf("YAML compilation failed: %v", err)
	}

	// Verify both formats contain the expected content
	if !strings.Contains(jsonOutput, `"name": "test"`) {
		t.Errorf("JSON output missing expected content")
	}

	if !strings.Contains(yamlOutput, "name: test") {
		t.Errorf("YAML output missing expected content")
	}

	t.Logf("JSON output:\n%s", jsonOutput)
	t.Logf("YAML output:\n%s", yamlOutput)
}

func TestEnvDirectiveYAML(t *testing.T) {
	// Set test environment variable
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	source := `
@brace "0.0.1"

@const "env" {
    TEST_VALUE = @env("TEST_VAR")
    DEFAULT_VALUE = @env("NONEXISTENT_VAR", "default")
}

test_value = :env.TEST_VALUE
default_value = :env.DEFAULT_VALUE
`

	compiler := NewWithFormat(transform.FormatYAML)
	output, err := compiler.Compile(source)
	if err != nil {
		t.Fatalf("YAML compilation failed: %v", err)
	}

	if !strings.Contains(output, "test_value: test_value") {
		t.Errorf("Expected environment variable to be resolved in YAML")
	}

	if !strings.Contains(output, "default_value: default") {
		t.Errorf("Expected default value to be used in YAML")
	}

	t.Logf("YAML environment test output:\n%s", output)
}
