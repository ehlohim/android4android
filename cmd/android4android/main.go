package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"android4android/internal/preflight"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, preflight.Check))
}

func run(args []string, stdout, stderr io.Writer, check func() preflight.Report) int {
	if len(args) == 0 || len(args) == 1 && (args[0] == "-h" || args[0] == "--help" || args[0] == "help") {
		fmt.Fprintln(stdout, "Usage: android4android doctor [--json]")
		return 0
	}
	if args[0] != "doctor" || len(args) > 2 || len(args) == 2 && args[1] != "--json" {
		fmt.Fprintln(stderr, "Usage: android4android doctor [--json]")
		return 64
	}
	report := check()
	if len(args) == 2 {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			fmt.Fprintln(stderr, "write diagnostics:", err)
			return 1
		}
	} else {
		fmt.Fprintf(stdout, "Host: %s/%s\n", report.OS, report.Arch)
		for _, result := range report.Checks {
			fmt.Fprintf(stdout, "%s %s: %s\n", result.Status, result.ID, result.Detail)
		}
	}
	if !report.Ready() {
		return 2
	}
	return 0
}
