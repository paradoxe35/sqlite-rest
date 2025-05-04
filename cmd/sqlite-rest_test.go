package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestVersion(t *testing.T) {
	// Test that VERSION is set
	if VERSION == "" {
		t.Errorf("VERSION should not be empty")
	}

	// Test version output format
	output := fmt.Sprintf("sqlite-rest version %s\n", VERSION)
	if !strings.Contains(output, "sqlite-rest version") {
		t.Errorf("Expected output to contain version, got: %s", output)
	}
}
