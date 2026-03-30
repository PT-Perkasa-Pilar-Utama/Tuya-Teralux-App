package orchestrator

import (
	"fmt"
	"regexp"
	"strings"
)

// FastIntentType represents the result of fast intent classification.
type FastIntentType string

const (
	FastIntentNone      FastIntentType = "none"
	FastIntentBlocked   FastIntentType = "blocked"
	FastIntentIdentity  FastIntentType = "identity"
	FastIntentControl   FastIntentType = "control"
	FastIntentDiscovery FastIntentType = "discovery"
)

// FastIntentResult contains the classification result and extracted control data.
type FastIntentResult struct {
	Intent       FastIntentType
	DeviceName   string  // extracted device name if control
	ActionType   string  // "on", "off", "brightness", "temperature", "fan_speed"
	Value        string  // extracted value (e.g., "50", "24", "level_2")
	ValuePercent int     // normalized percentage value if applicable
	Temperature  int     // temperature value if applicable
	Confidence   float64 // 0.0 to 1.0, how confident we are in this classification
}

// FastIntentRouter classifies prompts using deterministic rules.
type FastIntentRouter struct {
	// Compiled regex patterns for performance
	brightnessPattern  *regexp.Regexp
	temperaturePattern *regexp.Regexp
	fanSpeedPattern    *regexp.Regexp
	deviceNamePattern  *regexp.Regexp
}

// NewFastIntentRouter creates a new fast intent router with pre-compiled patterns.
func NewFastIntentRouter() *FastIntentRouter {
	return &FastIntentRouter{
		brightnessPattern:  regexp.MustCompile(`(?i)(\d+)\s*(persen|percent|%)|brightness\s*(\d+)|level\s*(\d+)`),
		temperaturePattern: regexp.MustCompile(`(?i)(\d+)\s*(derajat|degree|°c|celsius)|temp(?:erature)?\s*(\d+)`),
		fanSpeedPattern:    regexp.MustCompile(`(?i)(kipas|fan)\s*(level|speed|kecepatan)\s*(\d+)|fan\s*(low|medium|high)|kipas\s*(pelan|sedang|kencang)`),
		deviceNamePattern:  regexp.MustCompile(`(?i)(lampu|light|ac|kipas|fan|tv|speaker|perangkat|device)\s+([a-z0-9\s]+)`),
	}
}

// Classify analyzes a prompt and returns fast intent classification.
func (r *FastIntentRouter) Classify(prompt string) FastIntentResult {
	promptLower := strings.ToLower(strings.TrimSpace(prompt))

	// Check for identity prompts first
	if r.isIdentityPrompt(promptLower) {
		return FastIntentResult{
			Intent:     FastIntentIdentity,
			Confidence: 1.0,
		}
	}

	// Check for device discovery prompts
	if r.isDiscoveryPrompt(promptLower) {
		return FastIntentResult{
			Intent:     FastIntentDiscovery,
			Confidence: 1.0,
		}
	}

	// Check for control prompts
	if result, ok := r.isControlPrompt(promptLower); ok {
		return result
	}

	// No fast match
	return FastIntentResult{
		Intent:     FastIntentNone,
		Confidence: 0.0,
	}
}

// isIdentityPrompt checks if the prompt is asking about assistant identity.
func (r *FastIntentRouter) isIdentityPrompt(prompt string) bool {
	identityPatterns := []string{
		"kamu siapa",
		"siapa kamu",
		"who are you",
		"what are you",
		"nama kamu",
		"your name",
		"sensio",
		"asisten",
		"assistant",
	}

	for _, pattern := range identityPatterns {
		if strings.Contains(prompt, pattern) {
			// Make sure it's actually an identity question, not a command
			if !r.isCommandPrompt(prompt) {
				return true
			}
		}
	}
	return false
}

// isCommandPrompt checks if the prompt contains command verbs.
func (r *FastIntentRouter) isCommandPrompt(prompt string) bool {
	commandVerbs := []string{
		"nyalakan", "matikan", "hidupkan", "set", "atur",
		"turn on", "turn off", "set", "adjust",
	}

	for _, verb := range commandVerbs {
		if strings.Contains(prompt, verb) {
			return true
		}
	}
	return false
}

// isDiscoveryPrompt checks if the prompt is asking about available devices.
func (r *FastIntentRouter) isDiscoveryPrompt(prompt string) bool {
	discoveryPatterns := []string{
		"apa aja",
		"apa saja",
		"device apa",
		"perangkat apa",
		"lampu apa",
		"what devices",
		"which devices",
		"available devices",
		"connected devices",
		"bisa kontrol",
		"bisa control",
		"bisa saya kontrol",
		"bisa saya control",
		"daftar device",
		"daftar perangkat",
		"list device",
		"list perangkat",
	}

	for _, pattern := range discoveryPatterns {
		if strings.Contains(prompt, pattern) {
			return true
		}
	}
	return false
}

// isControlPrompt checks if the prompt is a device control command.
func (r *FastIntentRouter) isControlPrompt(prompt string) (FastIntentResult, bool) {
	// Check for on/off commands
	if strings.Contains(prompt, "nyalakan") || strings.Contains(prompt, "hidupkan") || strings.Contains(prompt, "turn on") {
		deviceName := r.extractDeviceName(prompt)
		if deviceName != "" {
			return FastIntentResult{
				Intent:     FastIntentControl,
				DeviceName: deviceName,
				ActionType: "on",
				Confidence: 0.9,
			}, true
		}
	}

	if strings.Contains(prompt, "matikan") || strings.Contains(prompt, "turn off") {
		deviceName := r.extractDeviceName(prompt)
		if deviceName != "" {
			return FastIntentResult{
				Intent:     FastIntentControl,
				DeviceName: deviceName,
				ActionType: "off",
				Confidence: 0.9,
			}, true
		}
	}

	// Check for brightness commands
	if strings.Contains(prompt, "brightness") || strings.Contains(prompt, "kecerahan") || strings.Contains(prompt, "persen") || strings.Contains(prompt, "percent") {
		deviceName := r.extractDeviceName(prompt)
		matches := r.brightnessPattern.FindStringSubmatch(prompt)
		if len(matches) > 1 {
			value := ""
			valuePercent := 0
			for i := 1; i < len(matches); i++ {
				if matches[i] != "" {
					value = matches[i]
					// Try to parse as percentage
					if i <= 2 {
						// First group might be percentage
						valuePercent = r.parsePercentage(matches[i])
					} else {
						// Later groups might be level
						valuePercent = r.parseLevel(matches[i])
					}
					break
				}
			}
			return FastIntentResult{
				Intent:       FastIntentControl,
				DeviceName:   deviceName,
				ActionType:   "brightness",
				Value:        value,
				ValuePercent: valuePercent,
				Confidence:   0.85,
			}, true
		}
	}

	// Check for temperature commands
	if strings.Contains(prompt, "suhu") || strings.Contains(prompt, "temperature") || strings.Contains(prompt, "derajat") || strings.Contains(prompt, "degree") {
		deviceName := r.extractDeviceName(prompt)
		if deviceName == "" {
			deviceName = "ac" // default to AC for temperature commands
		}
		matches := r.temperaturePattern.FindStringSubmatch(prompt)
		if len(matches) > 1 {
			value := ""
			temp := 0
			for i := 1; i < len(matches); i++ {
				if matches[i] != "" {
					value = matches[i]
					temp = r.parseInt(matches[i])
					break
				}
			}
			return FastIntentResult{
				Intent:      FastIntentControl,
				DeviceName:  deviceName,
				ActionType:  "temperature",
				Value:       value,
				Temperature: temp,
				Confidence:  0.85,
			}, true
		}
	}

	// Check for fan speed commands
	if strings.Contains(prompt, "kipas") || strings.Contains(prompt, "fan") {
		deviceName := r.extractDeviceName(prompt)
		if deviceName == "" {
			deviceName = "kipas"
		}
		matches := r.fanSpeedPattern.FindStringSubmatch(prompt)
		if len(matches) > 0 {
			value := ""
			valuePercent := 0
			// Check for level-based fan speed
			for i := 1; i < len(matches); i++ {
				if matches[i] != "" {
					value = matches[i]
					if r.isNumeric(matches[i]) {
						valuePercent = r.parseLevel(matches[i])
					} else {
						// Map text levels to percentages
						valuePercent = r.mapFanLevelToPercent(matches[i])
					}
					break
				}
			}
			return FastIntentResult{
				Intent:       FastIntentControl,
				DeviceName:   deviceName,
				ActionType:   "fan_speed",
				Value:        value,
				ValuePercent: valuePercent,
				Confidence:   0.8,
			}, true
		}

		// Simple on/off for fan
		if strings.Contains(prompt, "nyalakan") || strings.Contains(prompt, "hidupkan") {
			return FastIntentResult{
				Intent:     FastIntentControl,
				DeviceName: deviceName,
				ActionType: "on",
				Confidence: 0.85,
			}, true
		}
		if strings.Contains(prompt, "matikan") {
			return FastIntentResult{
				Intent:     FastIntentControl,
				DeviceName: deviceName,
				ActionType: "off",
				Confidence: 0.85,
			}, true
		}
	}

	return FastIntentResult{}, false
}

// extractDeviceName tries to extract a device name from the prompt.
func (r *FastIntentRouter) extractDeviceName(prompt string) string {
	// Common device names in Indonesian and English
	devices := []string{
		"lampu ruang tamu", "lampu kamar", "lampu tidur", "lampu dapur", "lampu mandi",
		"lampu teras", "lampu garasi", "lampu taman", "lampu meja",
		"ac ruang tamu", "ac kamar", "ac tidur", "ac dapur",
		"kipas ruang tamu", "kipas kamar", "kipas tidur",
		"tv ruang tamu", "tv kamar",
		"speaker ruang tamu", "speaker kamar",
		"light living room", "light bedroom", "light kitchen", "light bathroom",
		"ac living room", "ac bedroom",
		"fan living room", "fan bedroom",
	}

	promptLower := strings.ToLower(prompt)
	for _, device := range devices {
		if strings.Contains(promptLower, device) {
			return device
		}
	}

	// Try regex pattern match
	matches := r.deviceNamePattern.FindStringSubmatch(promptLower)
	if len(matches) > 2 {
		deviceType := matches[1]
		deviceLocation := strings.TrimSpace(matches[2])
		if deviceLocation != "" {
			return deviceType + " " + deviceLocation
		}
		return deviceType
	}

	// Check for simple device type references
	simpleDevices := []string{"lampu", "light", "ac", "kipas", "fan", "tv", "speaker"}
	for _, device := range simpleDevices {
		if strings.Contains(promptLower, device) {
			return device
		}
	}

	return ""
}

// parsePercentage tries to parse a string as a percentage value.
func (r *FastIntentRouter) parsePercentage(s string) int {
	val := r.parseInt(s)
	if val >= 0 && val <= 100 {
		return val
	}
	return 0
}

// parseLevel maps a level (1-5 typically) to a percentage.
func (r *FastIntentRouter) parseLevel(s string) int {
	val := r.parseInt(s)
	if val >= 1 && val <= 5 {
		// Map 1-5 to 20%, 40%, 60%, 80%, 100%
		return val * 20
	}
	if val >= 0 && val <= 100 {
		return val
	}
	return 0
}

// mapFanLevelToPercent maps text fan levels to percentages.
func (r *FastIntentRouter) mapFanLevelToPercent(level string) int {
	level = strings.ToLower(strings.TrimSpace(level))
	switch level {
	case "low", "pelan", "rendah", "1":
		return 33
	case "medium", "sedang", "2":
		return 66
	case "high", "kencang", "tinggi", "3":
		return 100
	default:
		return 0
	}
}

// parseInt tries to parse a string as an integer.
func (r *FastIntentRouter) parseInt(s string) int {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	if err != nil {
		return 0
	}
	return result
}

// isNumeric checks if a string is numeric.
func (r *FastIntentRouter) isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}
