package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"
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
		fmt.Printf("Error creating stdout pipe: %v\n", err)
		os.Exit(1)
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("Error starting command: %v\n", err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(stdout)

	// Stats
	var totalPassed, totalFailed, totalSkipped int
	pkgResults := make(map[string]*PackageResult)
	var failures []TestFailure
	testOutputs := make(map[string][]string) // Buffer output for each test to show on failure

	startTime := time.Now()

	fmt.Println(ColorBlue + ColorBold + "RUNNING TESTS..." + ColorReset)
	fmt.Println()

	for scanner.Scan() {
		line := scanner.Bytes()
		var event TestEvent
		if err := json.Unmarshal(line, &event); err != nil {
			// If not JSON (e.g. build failure output), print raw
			fmt.Println(string(line))
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
				// fmt.Printf("%s✓ %s%s\n", ColorGreen, event.Test, ColorReset) // Optional verbose
			} else if event.Package != "" {
				// Package Passed
				pkgResults[event.Package].Elapsed = event.Elapsed
				fmt.Printf("%s PASS %s %s(%.2fs)%s\n", ColorGreen, ColorReset, event.Package, event.Elapsed, ColorReset)
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

				fmt.Printf("%s✕ %s%s\n", ColorRed, event.Test, ColorReset)
			} else if event.Package != "" {
				// Package Failed
				pkgResults[event.Package].Elapsed = event.Elapsed
				fmt.Printf("%s FAIL %s %s(%.2fs)%s\n", ColorRed, ColorReset, event.Package, event.Elapsed, ColorReset)
			}
		case "skip":
			if event.Test != "" {
				totalSkipped++
				pkgResults[event.Package].Skipped++
				fmt.Printf("%s○ %s%s\n", ColorYellow, event.Test, ColorReset)
			}
		case "output":
			// We buffer output above, usually don't print immediately for cleaner "Jest-like" look
			// unless it's package level output not associated with a test
			if event.Test == "" && event.Output != "" {
				// Print build errors or package level logs
				// fmt.Print(event.Output)
			}
		}
	}

	cmd.Wait() // Wait for command to finish

	duration := time.Since(startTime)

	fmt.Println()

	// Print Failures Details
	if len(failures) > 0 {
		fmt.Println(ColorBold + "Summary of Failures:" + ColorReset)
		for _, f := range failures {
			fmt.Printf("%sFAIL: %s - %s%s\n", ColorRed, f.Package, f.Test, ColorReset)
			for _, line := range f.Output {
				// Indent output
				fmt.Print("    " + line)
			}
			fmt.Println()
		}
	}

	// Final Summary
	fmt.Println(ColorBold + "Test Suites:" + ColorReset + fmt.Sprintf(" %d passed, %d total", len(pkgResults), len(pkgResults))) // Simplified suites

	totalTests := totalPassed + totalFailed + totalSkipped

	fmt.Print(ColorBold + "Tests:       " + ColorReset)
	if totalFailed > 0 {
		fmt.Printf("%s%d failed%s, ", ColorRed, totalFailed, ColorReset)
	}
	if totalSkipped > 0 {
		fmt.Printf("%s%d skipped%s, ", ColorYellow, totalSkipped, ColorReset)
	}
	fmt.Printf("%s%d passed%s, %d total\n", ColorGreen, totalPassed, ColorReset, totalTests)

	fmt.Printf(ColorBold+"Time:        "+ColorReset+"%.2fs\n", duration.Seconds())

	if totalFailed > 0 {
		os.Exit(1)
	}
}
