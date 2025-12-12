package test_utils

import (
	"fmt"
	"testing"
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
)

// PrintTestHeader logs the test scenario header
func PrintTestHeader(t *testing.T, testName, scenario string) {
	msg := fmt.Sprintf("\n%sTest: %s (%s)%s", ColorBlue, testName, scenario, ColorReset)
	fmt.Println(msg)
}

// LogResult logs the final result in Green (Success) or Red (Failure)
func LogResult(t *testing.T, expected string, actual string, success bool) {
	fmt.Printf("\tExpected: %s\n", expected)
	if success {
		fmt.Printf("\t%sResult:   %s%s\n", ColorGreen, actual, ColorReset)
		t.Logf("\tExpected: %s | Result: %s (SUCCESS)", expected, actual)
	} else {
		fmt.Printf("\t%sResult:   %s%s\n", ColorRed, actual, ColorReset)
		t.Errorf("\tExpected: %s | Result: %s (FAILURE)", expected, actual)
		t.Fail()
	}
}

// AssertEqual is a helper that logs equality result
func AssertEqual(t *testing.T, expected interface{}, actual interface{}) {
	expStr := fmt.Sprintf("%v", expected)
	actStr := fmt.Sprintf("%v", actual)
	LogResult(t, expStr, actStr, expected == actual)
}

// AssertNotEqual helper
func AssertNotEqual(t *testing.T, notExpected interface{}, actual interface{}) {
	expStr := fmt.Sprintf("Not %v", notExpected)
	actStr := fmt.Sprintf("%v", actual)
	LogResult(t, expStr, actStr, notExpected != actual)
}
