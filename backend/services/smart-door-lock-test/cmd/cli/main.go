package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"sensio/backend/services/smart-door-lock-test/internal/config"
	"sensio/backend/services/smart-door-lock-test/internal/repository/tuya"
	"sensio/backend/services/smart-door-lock-test/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("❌ Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if err := cfg.Validate(); err != nil {
		fmt.Printf("❌ Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize repositories
	client := tuya.NewClient(cfg.Tuya.BaseURL, cfg.Tuya.ClientID, cfg.Tuya.AccessSecret)
	deviceRepo := tuya.NewDeviceRepository(client)
	commandRepo := tuya.NewCommandRepository(client)
	passwordRepo := tuya.NewPasswordRepository(client)

	// Initialize services
	deviceService := service.NewDeviceService(deviceRepo)
	commandService := service.NewCommandService(commandRepo)
	passwordService := service.NewPasswordService(passwordRepo)

	// Run CLI
	cli := NewCLI(deviceService, commandService, passwordService, cfg.Tuya.DeviceID)
	cli.Run()
}

// CLI handles command-line interaction
type CLI struct {
	deviceService    *service.DeviceService
	commandService   *service.CommandService
	passwordService  *service.PasswordService
	deviceID         string
	reader           *bufio.Reader
	currentDevice    interface{} // cache
	lastStatusRefresh time.Time
}

// NewCLI creates a new CLI handler
func NewCLI(
	deviceService *service.DeviceService,
	commandService *service.CommandService,
	passwordService *service.PasswordService,
	deviceID string,
) *CLI {
	return &CLI{
		deviceService:   deviceService,
		commandService:  commandService,
		passwordService: passwordService,
		deviceID:        deviceID,
		reader:          bufio.NewReader(os.Stdin),
	}
}

// Run starts the CLI main loop
func (c *CLI) Run() {
	c.printHeader()
	c.refreshDeviceStatus()

	for {
		c.printMenu()
		input := c.readInput()

		switch input {
		case "1":
			c.handleUnlock()
		case "2":
			c.handleLock()
		case "3":
			c.handleDynamicPassword()
		case "4":
			c.handleTemporaryPassword()
		case "5":
			c.refreshDeviceStatus()
		case "6":
			c.exit()
		default:
			fmt.Println("\n❌ Invalid option. Please try again.")
		}
	}
}

func (c *CLI) printHeader() {
	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║     Tuya Smart Door Lock - Control Panel              ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")
	fmt.Println()
}

func (c *CLI) printMenu() {
	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║                    Main Menu                           ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")
	fmt.Println("  1. 🔓 Unlock Door")
	fmt.Println("  2. 🔒 Lock Door")
	fmt.Println("  3. 🔑 Generate Dynamic Password (5 min)")
	fmt.Println("  4. 🔑 Generate Temporary Password")
	fmt.Println("  5. 🔄 Refresh Status")
	fmt.Println("  6. 🚪 Exit")
	fmt.Println()
	fmt.Print("  Select option: ")
}

func (c *CLI) readInput() string {
	input, _ := c.reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func (c *CLI) refreshDeviceStatus() {
	fmt.Println("\n📊 Device Status")
	fmt.Println("─────────────────────────────────────────────────────────")

	device, specs, err := c.deviceService.GetDevice(c.deviceID)
	if err != nil {
		fmt.Printf("❌ Failed to get device: %v\n", err)
		return
	}

	c.currentDevice = device
	c.lastStatusRefresh = time.Now()

	fmt.Printf("  Name:     %s\n", device.Name)
	fmt.Printf("  Category: %s\n", device.Category)
	fmt.Printf("  Online:   %v\n", c.onlineStatus(device.Online))
	fmt.Println()

	if len(device.Statuses) > 0 {
		fmt.Println("  Status:")
		for _, status := range device.Statuses {
			fmt.Printf("    %-22s: %v\n", status.Code, status.Value)
		}
	}

	if len(specs.Functions) > 0 {
		fmt.Println()
		fmt.Println("  Available Functions:")
		for _, fn := range specs.Functions {
			fmt.Printf("    %-22s (%s)\n", fn.Code, fn.Type)
		}
	}

	fmt.Println()
}

func (c *CLI) onlineStatus(online bool) string {
	if online {
		return "🟢 Online"
	}
	return "🔴 Offline"
}

func (c *CLI) handleUnlock() {
	fmt.Println("\n🔓 Unlocking door...")

	err := c.commandService.Unlock(c.deviceID)
	if err != nil {
		fmt.Printf("  ❌ Failed: %v\n", err)
		return
	}

	fmt.Println("  ✅ Door unlocked successfully!")
}

func (c *CLI) handleLock() {
	fmt.Println("\n🔒 Locking door...")

	err := c.commandService.Lock(c.deviceID)
	if err != nil {
		fmt.Printf("  ❌ Failed: %v\n", err)
		return
	}

	fmt.Println("  ✅ Door locked successfully!")
}

func (c *CLI) handleDynamicPassword() {
	fmt.Println("\n🔑 Generating dynamic password...")
	fmt.Println("   (One-time use, valid for 5 minutes)")
	fmt.Println()

	password, err := c.passwordService.GenerateDynamicPassword(c.deviceID)
	if err != nil {
		fmt.Printf("  ❌ Failed: %v\n", err)
		return
	}

	fmt.Printf("  ✅ Password: %s\n", password.Value)
	fmt.Printf("  ⏰ Valid until: %s\n", password.ExpireAt.Format("15:04:05"))
	fmt.Printf("  ⏱️  Duration: %d minutes\n", password.ValidMinutes)
	fmt.Println()
}

func (c *CLI) handleTemporaryPassword() {
	fmt.Println("\n🔑 Generate temporary password")
	fmt.Println()

	// Get duration
	fmt.Print("  Duration (minutes, default=60): ")
	durationInput, _ := c.reader.ReadString('\n')
	durationInput = strings.TrimSpace(durationInput)

	duration := 60
	if durationInput != "" {
		if d, err := strconv.Atoi(durationInput); err == nil && d > 0 {
			duration = d
		}
	}

	// Get custom password (optional)
	fmt.Print("  Custom password (leave empty for auto): ")
	customPwd, _ := c.reader.ReadString('\n')
	customPwd = strings.TrimSpace(customPwd)

	fmt.Println()
	fmt.Println("  Generating...")

	password, err := c.passwordService.GenerateTemporaryPassword(c.deviceID, duration, customPwd)
	if err != nil {
		fmt.Printf("  ❌ Failed: %v\n", err)
		return
	}

	fmt.Printf("  ✅ Password: %s\n", password.Value)
	fmt.Printf("  ⏰ Valid until: %s\n", password.ExpireAt.Format("15:04:05"))
	fmt.Printf("  ⏱️  Duration: %d minutes\n", password.ValidMinutes)
	fmt.Println()
}

func (c *CLI) exit() {
	fmt.Println("\n👋 Goodbye!")
	os.Exit(0)
}
