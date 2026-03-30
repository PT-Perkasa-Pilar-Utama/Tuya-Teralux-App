package main

import (
	"fmt"
	"os"
	"time"

	"sensio/backend/services/smart-door-lock-test/test/e2e/fixtures"
)

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║     Smart Door Lock - E2E Offline Tests               ║")
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
		fmt.Println("🟢 Device is ONLINE")
		fmt.Println()
		fmt.Println("⚠️  WARNING: Device should be OFFLINE for these tests!")
		fmt.Println()
		fmt.Println("Please disconnect your device from WiFi/power and:")
		fmt.Println("1. Wait 10 seconds for device to be marked offline")
		fmt.Println("2. Run this test again")
		fmt.Println()
		fmt.Print("Press Enter to continue anyway, or Ctrl+C to cancel...")
		fmt.Scanln()
	} else {
		fmt.Println("🔴 Device is OFFLINE - perfect for offline tests")
	}
	fmt.Println()

	// Run Phase 1: Offline Tests
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("PHASE 1: OFFLINE PASSWORD CREATION TESTS")
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println()

	runOfflineTests(helper)
}

func runOfflineTests(helper *fixtures.TestHelper) {
	results := make(map[string]interface{})

	// E2E-OFF-01: Dynamic Password (offline)
	fmt.Println("🧪 Test E2E-OFF-01: Create dynamic password (offline)")
	fmt.Println("   Goal: Understand Tuya API response when device offline")
	password, err := helper.PasswordService.GenerateDynamicPassword(helper.Config.DeviceID)
	if err != nil {
		fmt.Printf("   ⚠️  API returned error: %v\n", err)
		results["E2E-OFF-01"] = map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
			"note":   "Tuya rejects offline requests immediately",
		}
	} else {
		fmt.Printf("   ✅ API accepted request (may be pending sync)\n")
		fmt.Printf("      Password: %s\n", password.Value)
		fmt.Printf("      Type: %s\n", password.Type)
		fmt.Printf("      Valid until: %s\n", password.ExpireAt.Format(time.RFC3339))
		results["E2E-OFF-01"] = map[string]interface{}{
			"status":   "accepted",
			"password": password.Value,
			"note":     "Tuya may queue for later sync",
		}
	}
	fmt.Println()

	// E2E-OFF-02: Temporary Password (offline)
	fmt.Println("🧪 Test E2E-OFF-02: Create temporary password 60min (offline)")
	fmt.Println("   Goal: Understand Tuya API response for temp password when offline")
	password, err = helper.PasswordService.GenerateTemporaryPassword(helper.Config.DeviceID, 60, "")
	if err != nil {
		fmt.Printf("   ⚠️  API returned error: %v\n", err)
		results["E2E-OFF-02"] = map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
			"note":   "Tuya rejects offline requests immediately",
		}
	} else {
		fmt.Printf("   ✅ API accepted request (may be pending sync)\n")
		fmt.Printf("      Password: %s\n", password.Value)
		fmt.Printf("      Type: %s\n", password.Type)
		fmt.Printf("      Valid until: %s\n", password.ExpireAt.Format(time.RFC3339))
		results["E2E-OFF-02"] = map[string]interface{}{
			"status":   "accepted",
			"password": password.Value,
			"note":     "Tuya may queue for later sync",
		}
	}
	fmt.Println()

	// E2E-OFF-03: Custom Password (offline)
	fmt.Println("🧪 Test E2E-OFF-03: Create custom password '999999' (offline)")
	fmt.Println("   Goal: Test custom password value when device offline")
	password, err = helper.PasswordService.GenerateTemporaryPassword(helper.Config.DeviceID, 120, "999999")
	if err != nil {
		fmt.Printf("   ⚠️  API returned error: %v\n", err)
		results["E2E-OFF-03"] = map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
			"note":   "Tuya rejects offline requests immediately",
		}
	} else {
		fmt.Printf("   ✅ API accepted request (may be pending sync)\n")
		fmt.Printf("      Password: %s\n", password.Value)
		fmt.Printf("      Type: %s\n", password.Type)
		fmt.Printf("      Valid until: %s\n", password.ExpireAt.Format(time.RFC3339))
		results["E2E-OFF-03"] = map[string]interface{}{
			"status":   "accepted",
			"password": password.Value,
			"note":     "Tuya may queue for later sync",
		}
	}
	fmt.Println()

	// Summary
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("OFFLINE TEST SUMMARY")
	fmt.Println("═══════════════════════════════════════════════════════")
	accepted := 0
	errors := 0
	for testID, result := range results {
		resultMap := result.(map[string]interface{})
		if resultMap["status"] == "accepted" {
			accepted++
			fmt.Printf("✅ %s: ACCEPTED (password created)\n", testID)
		} else {
			errors++
			fmt.Printf("❌ %s: ERROR - %v\n", testID, resultMap["error"])
		}
	}
	fmt.Println()
	fmt.Printf("Total: %d accepted, %d errors\n", accepted, errors)
	fmt.Println()

	// Analysis
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("ANALYSIS")
	fmt.Println("═══════════════════════════════════════════════════════")
	if errors > 0 {
		fmt.Println("📝 FINDING: Tuya API rejects password creation when device is offline")
		fmt.Println()
		fmt.Println("IMPLICATION:")
		fmt.Println("- Smart Life offline password feature is NOT cloud-side deferred sync")
		fmt.Println("- Smart Life likely uses local/BLE gateway or device-side queue")
		fmt.Println("- To achieve parity, we need to implement our own pending queue")
		fmt.Println()
		fmt.Println("NEXT STEPS:")
		fmt.Println("1. Implement SQLite persistence for pending passwords")
		fmt.Println("2. Add worker to retry when device comes online")
		fmt.Println("3. Expose sync status in API responses")
	} else {
		fmt.Println("📝 FINDING: Tuya API accepts password creation even when device offline")
		fmt.Println()
		fmt.Println("IMPLICATION:")
		fmt.Println("- Tuya cloud handles deferred sync automatically")
		fmt.Println("- Smart Life parity may already work!")
		fmt.Println()
		fmt.Println("NEXT STEPS:")
		fmt.Println("1. Reconnect device to WiFi")
		fmt.Println("2. Test if passwords work on physical lock")
		fmt.Println("3. Verify sync delay")
	}
	fmt.Println()
}
