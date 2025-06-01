package compiler

import (
	"os"
	"strings"
	"testing"
)

func TestBasicCompilation(t *testing.T) {
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
		t.Fatalf("Compilation failed: %v", err)
	}

	// Check that output contains expected JSON structure
	if !strings.Contains(output, `"app_version": "1.0.0"`) {
		t.Errorf("Expected app_version to be resolved to '1.0.0'")
	}

	if !strings.Contains(output, `"database"`) {
		t.Errorf("Expected database table to be present")
	}

	t.Logf("Compilation output:\n%s", output)
}

func TestEnvDirective(t *testing.T) {
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

	compiler := New()
	output, err := compiler.Compile(source)
	if err != nil {
		t.Fatalf("Compilation failed: %v", err)
	}

	if !strings.Contains(output, `"test_value": "test_value"`) {
		t.Errorf("Expected environment variable to be resolved")
	}

	if !strings.Contains(output, `"default_value": "default"`) {
		t.Errorf("Expected default value to be used")
	}

	t.Logf("Environment test output:\n%s", output)
}
