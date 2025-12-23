// Copyright (c) Red Hat, Inc.
// Copyright Contributors to the Open Cluster Management project
// Licensed under the Apache License 2.0
package main

import (
	"fmt"
	"os"

	"github.com/ghodss/yaml"

	migration "github.com/dbuchanaRH/allowlist-migration-mcoa/pkg/migration"
	corev1 "k8s.io/api/core/v1"
)

/**
 * To run, use the following
 * go run tools/migrate_allowlist.go <path/to/allowlist.yaml> <path/to/write/dir/> <optional: targets.yaml>
 *
 * where targets.yaml is a file mounted in the prometheus or prometheusAgent pod
 */
func main() {
	if len(os.Args) < 3 || len(os.Args) > 4 {
		fmt.Fprintf(os.Stderr, "Usage: go run migrate_allowlist.go <path/to/allowlist.yaml> <path/to/write/dir/> <optional: targets.yaml>\n")
		os.Exit(1)
	}

	inputFile := os.Args[1]
	outputDir := os.Args[2]
	targets := "metrics_list.yaml"
	if len(os.Args) == 4 {
		targets = os.Args[3]
	}

	fmt.Fprintf(os.Stdout, "Using configmap %s\n", targets)

	// Read and parse the allowlist file
	allowFile, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read file %s: %v\n", inputFile, err)
		os.Exit(1)
	}

	var cm corev1.ConfigMap
	if err := yaml.Unmarshal(allowFile, &cm); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to unmarshal ConfigMap YAML: %v\n", err)
		os.Exit(1)
	}

	// Migrate allowlist to ScrapeConfig and PrometheusRule
	scrapeConfig, prometheusRule, err := migration.MigrateAllowlist(cm, targets)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to migrate allowlist: %v\n", err)
		os.Exit(1)
	}

	scrapeConfigName := scrapeConfig.ObjectMeta.Name
	prometheusRuleName := prometheusRule.ObjectMeta.Name
	// Marshal to YAML
	scrapeConfigYAML, err := yaml.Marshal(scrapeConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal scrapeConfig: %v\n", err)
		os.Exit(1)
	}

	prometheusRuleYAML, err := yaml.Marshal(prometheusRule)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal prometheusRule: %v\n", err)
		os.Exit(1)
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create output directory %s: %v\n", outputDir, err)
		os.Exit(1)
	}

	// Write files
	scrapeConfigPath := fmt.Sprintf("%s/%s.yaml", outputDir, scrapeConfigName)
	if err := os.WriteFile(scrapeConfigPath, scrapeConfigYAML, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write %s: %v\n", scrapeConfigPath, err)
		os.Exit(1)
	}

	prometheusRulePath := fmt.Sprintf("%s/%s.yaml", outputDir, prometheusRuleName)
	if err := os.WriteFile(prometheusRulePath, prometheusRuleYAML, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write %s: %v\n", prometheusRulePath, err)
		os.Exit(1)
	}

	fmt.Printf("Successfully migrated allowlist:\n")
	fmt.Printf("  - %s\n", scrapeConfigPath)
	fmt.Printf("  - %s\n", prometheusRulePath)
}
