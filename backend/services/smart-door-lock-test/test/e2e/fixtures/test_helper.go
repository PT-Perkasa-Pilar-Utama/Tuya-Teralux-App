package fixtures

import (
	"fmt"
	"time"

	"sensio/backend/services/smart-door-lock-test/internal/config"
	"sensio/backend/services/smart-door-lock-test/internal/repository/tuya"
	"sensio/backend/services/smart-door-lock-test/internal/service"
)

// TestHelper provides common E2E test utilities
type TestHelper struct {
	Config          *TestConfig
	DeviceService   *service.DeviceService
	PasswordService *service.PasswordService
	CommandService  *service.CommandService
	TuyaClient      *tuya.Client
}

// NewTestHelper creates a new test helper
func NewTestHelper() (*TestHelper, error) {
	testConfig, err := LoadTestConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load test config: %w", err)
	}

	// Load app config
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load app config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Initialize Tuya client
	client := tuya.NewClient(cfg.Tuya.BaseURL, cfg.Tuya.ClientID, cfg.Tuya.AccessSecret)

	// Initialize repositories
	deviceRepo := tuya.NewDeviceRepository(client)
	commandRepo := tuya.NewCommandRepository(client)
	passwordRepo := tuya.NewPasswordRepository(client)

	// Initialize services
	deviceService := service.NewDeviceService(deviceRepo)
	commandService := service.NewCommandService(commandRepo)
	passwordService := service.NewPasswordService(passwordRepo)

	return &TestHelper{
		Config:          testConfig,
		DeviceService:   deviceService,
		PasswordService: passwordService,
		CommandService:  commandService,
		TuyaClient:      client,
	}, nil
}

// CheckDeviceOnline checks if the device is currently online
func (h *TestHelper) CheckDeviceOnline() (bool, error) {
	return h.DeviceService.IsOnline(h.Config.DeviceID)
}

// WaitForDeviceOnline waits for device to come online with timeout
func (h *TestHelper) WaitForDeviceOnline(timeout time.Duration) (bool, error) {
	start := time.Now()
	for time.Since(start) < timeout {
		online, err := h.CheckDeviceOnline()
		if err != nil {
			return false, err
		}
		if online {
			return true, nil
		}
		time.Sleep(2 * time.Second)
	}
	return false, nil
}

// GenerateTestReport creates a formatted test report
func (h *TestHelper) GenerateTestReport(testID, testName, status string, details map[string]interface{}) string {
	report := fmt.Sprintf("╔════════════════════════════════════════════════════════╗\n")
	report += fmt.Sprintf("║ TEST: %-54s ║\n", testID)
	report += fmt.Sprintf("╠════════════════════════════════════════════════════════╣\n")
	report += fmt.Sprintf("║ Name:   %-50s ║\n", testName)
	report += fmt.Sprintf("║ Status: %-50s ║\n", status)
	report += fmt.Sprintf("╠════════════════════════════════════════════════════════╣\n")

	for key, value := range details {
		line := fmt.Sprintf("║ %-8s: %-47s ║", key, fmt.Sprintf("%v", value))
		report += line + "\n"
	}

	report += fmt.Sprintf("╚════════════════════════════════════════════════════════╝\n")
	return report
}
