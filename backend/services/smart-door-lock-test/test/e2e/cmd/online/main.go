package main

import (
	"fmt"
	"os"
	"time"

	"sensio/backend/services/smart-door-lock-test/test/e2e/fixtures"
)

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║     Smart Door Lock - E2E Test Runner                 ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Initialize test helper
	helper, err := fixtures.NewTestHelper()
	if err != nil {
		fmt.Printf("❌ Failed to initialize test helper: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Test helper initialized")
	fmt.Println()

	// Check device status first
	fmt.Println("📊 Checking device status...")
	online, err := helper.CheckDeviceOnline()
	if err != nil {
		fmt.Printf("❌ Failed to check device status: %v\n", err)
		os.Exit(1)
	}

	if online {
		fmt.Println("🟢 Device is ONLINE - proceeding with online tests")
		fmt.Println()
	} else {
		fmt.Println("🔴 Device is OFFLINE")
		fmt.Println()
		fmt.Println("═══════════════════════════════════════════════════════")
		fmt.Println("WAITING FOR DEVICE TO COME ONLINE")
		fmt.Println("═══════════════════════════════════════════════════════")
		fmt.Println()
		fmt.Println("⏱️  Timeout: 120 seconds")
		fmt.Println()
		fmt.Println("Please connect your device to WiFi now...")
		fmt.Println()

		// Wait for device with timeout
		timeout := 120 * time.Second
		interval := 5 * time.Second

		start := time.Now()
		for time.Since(start) < timeout {
			remaining := timeout - time.Since(start)
			fmt.Printf("⏳ Waiting for device... (%.0fs remaining)\n", remaining.Seconds())

			online, err = helper.CheckDeviceOnline()
			if err != nil {
				fmt.Printf("   Error checking status: %v\n", err)
			} else if online {
				fmt.Println()
				fmt.Println("🟢 Device is now ONLINE!")
				fmt.Println()
				break
			}

			time.Sleep(interval)
		}

		if !online {
			fmt.Println()
			fmt.Println("❌ TIMEOUT: Device did not come online within 120 seconds")
			fmt.Println()
			fmt.Println("Options:")
			fmt.Println("1. Check device power and WiFi connection")
			fmt.Println("2. Verify device in Smart Life app")
			fmt.Println("3. Try again")
			fmt.Println()
			fmt.Println("Proceeding with tests anyway (will likely fail)...")
			fmt.Println()
		}
	}

	// Run Phase 1: Online Tests
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("PHASE 1: ONLINE PASSWORD CREATION TESTS")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println()

	runOnlineTests(helper)

	// Prompt for offline tests
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("NEXT: OFFLINE TESTS")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("To run offline tests:")
	fmt.Println("1. Disconnect your device from WiFi/power")
	fmt.Println("2. Wait 10 seconds")
	fmt.Println("3. Run: go run test/e2e/cmd/offline/main.go")
	fmt.Println()
}

func runOnlineTests(helper *fixtures.TestHelper) {
	results := make(map[string]string)

	// E2E-ON-01: Dynamic Password
	fmt.Println("🧪 Test E2E-ON-01: Create dynamic password (online)")
	password, err := helper.PasswordService.GenerateDynamicPassword(helper.Config.DeviceID)
	if err != nil {
		fmt.Printf("   ❌ FAILED: %v\n", err)
		results["E2E-ON-01"] = "FAILED"
	} else {
		fmt.Printf("   ✅ PASSED\n")
		fmt.Printf("      Password: %s\n", password.Value)
		fmt.Printf("      Type: %s\n", password.Type)
		fmt.Printf("      Valid until: %s\n", password.ExpireAt.Format(time.RFC3339))
		fmt.Printf("      Duration: %d minutes\n", password.ValidMinutes)
		results["E2E-ON-01"] = "PASSED"
	}
	fmt.Println()

	// E2E-ON-02: Temporary Password (60 min)
	fmt.Println("🧪 Test E2E-ON-02: Create temporary password 60min (online)")
	password, err = helper.PasswordService.GenerateTemporaryPassword(helper.Config.DeviceID, 60, "")
	if err != nil {
		fmt.Printf("   ❌ FAILED: %v\n", err)
		results["E2E-ON-02"] = "FAILED"
	} else {
		fmt.Printf("   ✅ PASSED\n")
		fmt.Printf("      Password: %s\n", password.Value)
		fmt.Printf("      Type: %s\n", password.Type)
		fmt.Printf("      Valid until: %s\n", password.ExpireAt.Format(time.RFC3339))
		fmt.Printf("      Duration: %d minutes\n", password.ValidMinutes)
		results["E2E-ON-02"] = "PASSED"
	}
	fmt.Println()

	// E2E-ON-03: Custom Temporary Password
	fmt.Println("🧪 Test E2E-ON-03: Create custom temporary password (online)")
	password, err = helper.PasswordService.GenerateTemporaryPassword(helper.Config.DeviceID, 120, "123456")
	if err != nil {
		fmt.Printf("   ❌ FAILED: %v\n", err)
		results["E2E-ON-03"] = "FAILED"
	} else {
		fmt.Printf("   ✅ PASSED\n")
		fmt.Printf("      Password: %s\n", password.Value)
		fmt.Printf("      Type: %s\n", password.Type)
		fmt.Printf("      Valid until: %s\n", password.ExpireAt.Format(time.RFC3339))
		fmt.Printf("      Duration: %d minutes\n", password.ValidMinutes)
		results["E2E-ON-03"] = "PASSED"
	}
	fmt.Println()

	// E2E-ON-04: Long Duration Password
	fmt.Println("🧪 Test E2E-ON-04: Create long-duration password 1year (online)")
	password, err = helper.PasswordService.GenerateTemporaryPassword(helper.Config.DeviceID, 525600, "")
	if err != nil {
		fmt.Printf("   ❌ FAILED: %v\n", err)
		results["E2E-ON-04"] = "FAILED"
	} else {
		fmt.Printf("   ✅ PASSED\n")
		fmt.Printf("      Password: %s\n", password.Value)
		fmt.Printf("      Type: %s\n", password.Type)
		fmt.Printf("      Valid until: %s\n", password.ExpireAt.Format(time.RFC3339))
		fmt.Printf("      Duration: %d minutes (%.1f days)\n", password.ValidMinutes, float64(password.ValidMinutes)/1440.0)
		results["E2E-ON-04"] = "PASSED"
	}
	fmt.Println()

	// Summary
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("TEST SUMMARY")
	fmt.Println("═══════════════════════════════════════════════════════")
	passed := 0
	failed := 0
	for testID, result := range results {
		if result == "PASSED" {
			passed++
			fmt.Printf("✅ %s: %s\n", testID, result)
		} else {
			failed++
			fmt.Printf("❌ %s: %s\n", testID, result)
		}
	}
	fmt.Println()
	fmt.Printf("Total: %d passed, %d failed\n", passed, failed)
	fmt.Println()
}
