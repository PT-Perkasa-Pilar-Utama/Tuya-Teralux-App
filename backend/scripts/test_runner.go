package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"teralux_app/domain/common/utils"
)

type TestEvent struct {
	Time    time.Time `json:"Time"`
	Action  string    `json:"Action"`
	Package string    `json:"Package"`
	Test    string    `json:"Test"`
	Elapsed float64   `json:"Elapsed"`
	Output  string    `json:"Output"`
}

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorGray   = "\033[90m"
	ColorBold   = "\033[1m"
)

type PackageResult struct {
	Name    string
	Elapsed float64
	Failed  int
	Passed  int
	Skipped int
}

type TestFailure struct {
	Package string
	Test    string
	Output  []string
}

func main() {
	cmd := exec.Command("go", "test", "-json", "./...")
	// Force color output from tests themselves if they support it, though -json usually strips it.
	// We rely on our own coloring.

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		utils.LogError("Error creating stdout pipe: %v", err)
		os.Exit(1)
	}

	if err := cmd.Start(); err != nil {
		utils.LogError("Error starting command: %v", err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(stdout)

	// Stats
	var totalPassed, totalFailed, totalSkipped int
	pkgResults := make(map[string]*PackageResult)
	var failures []TestFailure
	testOutputs := make(map[string][]string) // Buffer output for each test to show on failure

	startTime := time.Now()

	utils.LogInfo(ColorBlue + ColorBold + "RUNNING TESTS..." + ColorReset)
	utils.LogInfo("")

	for scanner.Scan() {
		line := scanner.Bytes()
		var event TestEvent
		if err := json.Unmarshal(line, &event); err != nil {
			// If not JSON (e.g. build failure output), print raw
			utils.LogInfo("%s", string(line))
			continue
		}

		// Track Package Results
		if event.Package != "" {
			if _, exists := pkgResults[event.Package]; !exists {
				pkgResults[event.Package] = &PackageResult{Name: event.Package}
			}
		}

		// Buffer output for tests (to display context on failure)
		if event.Test != "" && event.Output != "" {
			key := fmt.Sprintf("%s|%s", event.Package, event.Test)
			testOutputs[key] = append(testOutputs[key], event.Output)
		}

		switch event.Action {
		case "pass":
			if event.Test != "" {
				// Individual Test Passed
				totalPassed++
				pkgResults[event.Package].Passed++
				// utils.LogInfo("%s✓ %s%s", ColorGreen, event.Test, ColorReset) // Optional verbose
			} else if event.Package != "" {
				// Package Passed
				pkgResults[event.Package].Elapsed = event.Elapsed
				utils.LogInfo("%s PASS %s %s(%.2fs)%s", ColorGreen, ColorReset, event.Package, event.Elapsed, ColorReset)
			}
		case "fail":
			if event.Test != "" {
				// Individual Test Failed
				totalFailed++
				pkgResults[event.Package].Failed++

				key := fmt.Sprintf("%s|%s", event.Package, event.Test)
				failures = append(failures, TestFailure{
					Package: event.Package,
					Test:    event.Test,
					Output:  testOutputs[key],
				})

				utils.LogError("%s✕ %s%s", ColorRed, event.Test, ColorReset)
			} else if event.Package != "" {
				// Package Failed
				pkgResults[event.Package].Elapsed = event.Elapsed
				utils.LogError("%s FAIL %s %s(%.2fs)%s", ColorRed, ColorReset, event.Package, event.Elapsed, ColorReset)
			}
		case "skip":
			if event.Test != "" {
				totalSkipped++
				pkgResults[event.Package].Skipped++
				utils.LogInfo("%s○ %s%s", ColorYellow, event.Test, ColorReset)
			}
		case "output":
			// We buffer output above, usually don't print immediately for cleaner "Jest-like" look
			// unless it's package level output not associated with a test
			if event.Test == "" && event.Output != "" {
				// Print build errors or package level logs
				// utils.LogInfo("%s", event.Output)
			}
		}
	}

	cmd.Wait() // Wait for command to finish

	duration := time.Since(startTime)

	utils.LogInfo("")

	// Print Failures Details
	if len(failures) > 0 {
		utils.LogError(ColorBold + "Summary of Failures:" + ColorReset)
		for _, f := range failures {
			utils.LogError("%sFAIL: %s - %s%s", ColorRed, f.Package, f.Test, ColorReset)
			for _, line := range f.Output {
				// Indent output
				utils.LogError("    %s", line)
			}
			utils.LogInfo("")
		}
	}

	// Final Summary
	totalTests := totalPassed + totalFailed + totalSkipped

	// Build and print a single summary line to keep colors intact
	summary := ColorBold + "Tests:       " + ColorReset
	if totalFailed > 0 {
		summary += fmt.Sprintf("%s%d failed%s, ", ColorRed, totalFailed, ColorReset)
	}
	if totalSkipped > 0 {
		summary += fmt.Sprintf("%s%d skipped%s, ", ColorYellow, totalSkipped, ColorReset)
	}
	summary += fmt.Sprintf("%s%d passed%s, %d total", ColorGreen, totalPassed, ColorReset, totalTests)
	utils.LogInfo("%s", summary)

	utils.LogInfo("%sTime:        %s%.2fs", ColorBold, ColorReset, duration.Seconds())

	if totalFailed > 0 {
		os.Exit(1)
	}
}
