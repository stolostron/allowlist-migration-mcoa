// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project
// Licensed under the Apache License 2.0
package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// testAllowlistYAML is a valid ConfigMap YAML for testing
const testAllowlistYAML = `apiVersion: v1
kind: ConfigMap
metadata:
  name: observability-metrics-custom-allowlist
  namespace: open-cluster-management-observability
data:
  metrics_list.yaml: |
    names:
      - up
      - container_cpu_usage_seconds_total
    matches:
      - __name__="container_memory_cache",container!=""
    recording_rules:
      - record: container_memory_rss:sum
        expr: sum(container_memory_rss) by (container, namespace)
`

// buildTestBinary builds the main binary for testing and returns its path
func buildTestBinary(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "migrate_allowlist")

	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Dir = filepath.Dir(binaryPath)
	// Build from the source directory
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build test binary: %v\nOutput: %s", err, output)
	}
	return binaryPath
}

// TestValidInputProducesOutputFiles verifies that with valid input, main outputs
// both scrapeconfig and prometheusrule files to the desired path
func TestValidInputProducesOutputFiles(t *testing.T) {
	binaryPath := buildTestBinary(t)

	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Create input allowlist file
	inputFile := filepath.Join(tmpDir, "observability-metrics-custom-allowlist.yaml")
	if err := os.WriteFile(inputFile, []byte(testAllowlistYAML), 0644); err != nil {
		t.Fatalf("failed to create input file: %v", err)
	}

	// Create output directory
	outputDir := filepath.Join(tmpDir, "output")

	// Test with 3 arguments (input, output, using default targets)
	t.Run("three arguments - valid input", func(t *testing.T) {
		// Clean output directory for this test
		testOutputDir := filepath.Join(outputDir, "three_args")

		cmd := exec.Command(binaryPath, inputFile, testOutputDir)
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("command failed: %v\nOutput: %s", err, output)
		}

		// Verify scrapeconfig file was created
		scrapeConfigPath := filepath.Join(testOutputDir, "custom-scrapeconfig.yaml")
		if _, err := os.Stat(scrapeConfigPath); os.IsNotExist(err) {
			t.Errorf("expected scrapeconfig file to exist at %s", scrapeConfigPath)
		}

		// Verify prometheusrule file was created
		prometheusRulePath := filepath.Join(testOutputDir, "custom-prometheusrule.yaml")
		if _, err := os.Stat(prometheusRulePath); os.IsNotExist(err) {
			t.Errorf("expected prometheusrule file to exist at %s", prometheusRulePath)
		}

		// Verify scrapeconfig content contains expected fields
		scrapeConfigContent, err := os.ReadFile(scrapeConfigPath)
		if err != nil {
			t.Fatalf("failed to read scrapeconfig file: %v", err)
		}
		if !strings.Contains(string(scrapeConfigContent), "kind: ScrapeConfig") {
			t.Errorf("scrapeconfig file does not contain 'kind: ScrapeConfig'")
		}

		// Verify prometheusrule content contains expected fields
		prometheusRuleContent, err := os.ReadFile(prometheusRulePath)
		if err != nil {
			t.Fatalf("failed to read prometheusrule file: %v", err)
		}
		if !strings.Contains(string(prometheusRuleContent), "kind: PrometheusRule") {
			t.Errorf("prometheusrule file does not contain 'kind: PrometheusRule'")
		}
	})

	// Test with 4 arguments (input, output, custom targets)
	t.Run("four arguments - valid input with custom targets", func(t *testing.T) {
		// Create input with custom targets key
		customAllowlistYAML := `apiVersion: v1
kind: ConfigMap
metadata:
  name: observability-metrics-custom-allowlist
  namespace: open-cluster-management-observability
data:
  custom_targets.yaml: |
    names:
      - up
    matches:
      - __name__="test_metric"
    recording_rules:
      - record: test_rule
        expr: sum(test_metric)
`
		customInputFile := filepath.Join(tmpDir, "observability-metrics-custom-allowlist.yaml")
		if err := os.WriteFile(customInputFile, []byte(customAllowlistYAML), 0644); err != nil {
			t.Fatalf("failed to create custom input file: %v", err)
		}

		testOutputDir := filepath.Join(outputDir, "four_args")

		cmd := exec.Command(binaryPath, customInputFile, testOutputDir, "custom_targets.yaml")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("command failed: %v\nOutput: %s", err, output)
		}

		// Verify scrapeconfig file was created
		scrapeConfigPath := filepath.Join(testOutputDir, "custom-scrapeconfig.yaml")
		if _, err := os.Stat(scrapeConfigPath); os.IsNotExist(err) {
			t.Errorf("expected scrapeconfig file to exist at %s", scrapeConfigPath)
		}

		// Verify prometheusrule file was created
		prometheusRulePath := filepath.Join(testOutputDir, "custom-prometheusrule.yaml")
		if _, err := os.Stat(prometheusRulePath); os.IsNotExist(err) {
			t.Errorf("expected prometheusrule file to exist at %s", prometheusRulePath)
		}
	})
}
