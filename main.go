// Package main is the entry point for the cloud-init to Butane transpiler CLI.
//
// Usage:
//
//	transpiler -input cloud-config.yaml -output butane.yaml
//	transpiler -input cloud-config.yaml   # outputs to stdout
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sakshamgupta/flatcar-butane-transpiler/pkg/cloudconfig"
	"github.com/sakshamgupta/flatcar-butane-transpiler/pkg/transpiler"
	"gopkg.in/yaml.v3"
)

func main() {
	inputPath := flag.String("input", "", "Path to the cloud-config YAML file (required)")
	outputPath := flag.String("output", "", "Path to write Butane YAML output (default: stdout)")
	strict := flag.Bool("strict", false, "Exit with error if any unsupported fields are encountered")
	flag.Parse()

	if *inputPath == "" {
		fmt.Fprintln(os.Stderr, "Error: -input flag is required")
		flag.Usage()
		os.Exit(1)
	}

	// Parse cloud-config
	cfg, err := cloudconfig.Parse(*inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing cloud-config: %v\n", err)
		os.Exit(1)
	}

	// Transpile to Butane
	butaneCfg, warnings, err := transpiler.Transpile(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error transpiling config: %v\n", err)
		os.Exit(1)
	}

	// Print warnings to stderr
	for _, w := range warnings {
		fmt.Fprintf(os.Stderr, "WARNING: %s\n", w)
	}

	if *strict && len(warnings) > 0 {
		fmt.Fprintln(os.Stderr, "Exiting due to warnings in strict mode")
		os.Exit(1)
	}

	// Encode output as YAML
	out, err := yaml.Marshal(butaneCfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding Butane YAML: %v\n", err)
		os.Exit(1)
	}

	// Write to file or stdout
	if *outputPath == "" {
		fmt.Print(string(out))
	} else {
		if err := os.WriteFile(*outputPath, out, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Butane config written to %s\n", *outputPath)
	}
}
