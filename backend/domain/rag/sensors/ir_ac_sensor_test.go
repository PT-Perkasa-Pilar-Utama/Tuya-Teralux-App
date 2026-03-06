package sensors

import (
	tuyaDtos "sensio/domain/tuya/dtos"
	"testing"

	"github.com/stretchr/testify/mock"
)

type MockTuyaExecutor struct {
	mock.Mock
}

func (m *MockTuyaExecutor) SendIRACCommand(token string, deviceID string, remoteID string, params map[string]int) (bool, error) {
	args := m.Called(token, deviceID, remoteID, params)
	return args.Bool(0), args.Error(1)
}

func (m *MockTuyaExecutor) SendSwitchCommand(token string, deviceID string, commands []tuyaDtos.TuyaCommandDTO) (bool, error) {
	return false, nil
}

func TestIRACsensor_ExecuteControl(t *testing.T) {
	s := NewIRACsensor()
	token := "test-token"
	device := &tuyaDtos.TuyaDeviceDTO{
		ID:       "dev123",
		RemoteID: "rem456",
		Name:     "AC Rumah",
	}

	tests := []struct {
		name           string
		prompt         string
		history        []string
		expectedParams map[string]int
		expectedMsg    string
		expectedStatus int
		mockSuccess    bool
	}{
		{
			name:   "Standard temperature parse (25.)",
			prompt: "Ubah temperatur AC rumah ke 25.",
			expectedParams: map[string]int{
				"power": 1,
				"mode":  0,
				"temp":  25,
				"wind":  0,
			},
			expectedMsg:    "Successfully set mode to Cool, set temperature to 25°C, set fan speed to Auto AC Rumah.",
			expectedStatus: 0,
			mockSuccess:    true,
		},
		{
			name:   "Temperature with symbol (25°C)",
			prompt: "set suhu 25°C",
			expectedParams: map[string]int{
				"power": 1,
				"mode":  0,
				"temp":  25,
				"wind":  0,
			},
			expectedMsg:    "Successfully set mode to Cool, set temperature to 25°C, set fan speed to Auto AC Rumah.",
			expectedStatus: 0,
			mockSuccess:    true,
		},
		{
			name:   "Temperature with word (25 derajat)",
			prompt: "set temperature ke 25 derajat",
			expectedParams: map[string]int{
				"power": 1,
				"mode":  0,
				"temp":  25,
				"wind":  0,
			},
			expectedMsg:    "Successfully set mode to Cool, set temperature to 25°C, set fan speed to Auto AC Rumah.",
			expectedStatus: 0,
			mockSuccess:    true,
		},
		{
			name:   "Multiple temperatures (pick last)",
			prompt: "turunkan dari 18 ke 25",
			expectedParams: map[string]int{
				"power": 1,
				"mode":  0,
				"temp":  25,
				"wind":  0,
			},
			expectedMsg:    "Successfully set mode to Cool, set temperature to 25°C, set fan speed to Auto AC Rumah.",
			expectedStatus: 0,
			mockSuccess:    true,
		},
		{
			name:           "Invalid out-of-range (high)",
			prompt:         "set suhu 35",
			expectedStatus: 400,
			expectedMsg:    "Temperature 35 is out of range. Valid range is 16-30°C.",
		},
		{
			name:           "Invalid out-of-range (low)",
			prompt:         "set suhu 10",
			expectedStatus: 400,
			expectedMsg:    "Temperature 10 is out of range. Valid range is 16-30°C.",
		},
		{
			name:   "Mode override (Auto -> Cool if temp provided)",
			prompt: "set 25 mode auto",
			expectedParams: map[string]int{
				"power": 1,
				"mode":  0, // Overridden from 2 to 0
				"temp":  25,
				"wind":  0,
			},
			expectedMsg:    "Successfully set mode to Cool, set temperature to 25°C, set fan speed to Auto AC Rumah.",
			expectedStatus: 0,
			mockSuccess:    true,
		},
		{
			name:   "Mode override (Fan -> Cool if temp provided)",
			prompt: "set 25 mode fan",
			expectedParams: map[string]int{
				"power": 1,
				"mode":  0, // Overridden from 3 to 0
				"temp":  25,
				"wind":  0,
			},
			expectedMsg:    "Successfully set mode to Cool, set temperature to 25°C, set fan speed to Auto AC Rumah.",
			expectedStatus: 0,
			mockSuccess:    true,
		},
		{
			name:   "Fallback if no temp (Cool mode)",
			prompt: "nyalakan AC",
			expectedParams: map[string]int{
				"power": 1,
				"mode":  0,
				"temp":  18, // Default
				"wind":  0,
			},
			expectedMsg:    "Successfully set mode to Cool, set temperature to 18°C, set fan speed to Auto AC Rumah.",
			expectedStatus: 0,
			mockSuccess:    true,
		},
		{
			name:   "Context-aware: ignore duration (set suhu 25 selama 30 menit)",
			prompt: "set suhu 25 selama 30 menit",
			expectedParams: map[string]int{
				"power": 1,
				"mode":  0,
				"temp":  25,
				"wind":  0,
			},
			expectedMsg:    "Successfully set mode to Cool, set temperature to 25°C, set fan speed to Auto AC Rumah.",
			expectedStatus: 0,
			mockSuccess:    true,
		},
		{
			name:   "Context-aware: ignore device number (set AC 2 ke 25)",
			prompt: "set AC 2 ke 25",
			expectedParams: map[string]int{
				"power": 1,
				"mode":  0,
				"temp":  25,
				"wind":  0,
			},
			expectedMsg:    "Successfully set mode to Cool, set temperature to 25°C, set fan speed to Auto AC Rumah.",
			expectedStatus: 0,
			mockSuccess:    true,
		},
		{
			name:   "Context-aware: ignore fan speed level (mode fan 3)",
			prompt: "mode fan 3",
			expectedParams: map[string]int{
				"power": 1,
				"mode":  3,
				"wind":  0, // Default Auto for Fan mode if not specified as Low/Med/High word
			},
			expectedMsg:    "Successfully set mode to Wind, set fan speed to Auto AC Rumah.",
			expectedStatus: 0,
			mockSuccess:    true,
		},
		{
			name:   "Context-aware: fan speed level + temperature (set kipas 3 suhu 24)",
			prompt: "set kipas 3 suhu 24",
			expectedParams: map[string]int{
				"power": 1,
				"mode":  0, // Overridden to Cool because temp is present
				"temp":  24,
				"wind":  0,
			},
			expectedMsg:    "Successfully set mode to Cool, set temperature to 24°C, set fan speed to Auto AC Rumah.",
			expectedStatus: 0,
			mockSuccess:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExecutor := new(MockTuyaExecutor)
			if tt.expectedStatus == 0 || tt.expectedStatus == 200 {
				mockExecutor.On("SendIRACCommand", token, device.ID, device.RemoteID, tt.expectedParams).Return(tt.mockSuccess, nil)
			}

			result, err := s.ExecuteControl(token, device, tt.prompt, tt.history, mockExecutor)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.expectedStatus != 0 && result.HTTPStatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, result.HTTPStatusCode)
			}

			if result.Message != tt.expectedMsg {
				t.Errorf("expected message '%s', got '%s'", tt.expectedMsg, result.Message)
			}

			if tt.expectedStatus == 0 || tt.expectedStatus == 200 {
				mockExecutor.AssertExpectations(t)
			}
		})
	}
}

func TestIRACsensor_parseTemperature(t *testing.T) {
	s := &IRACsensor{}
	tests := []struct {
		prompt string
		want   int
		found  bool
	}{
		{"25", 25, true},
		{"25.", 25, true},
		{"25°C", 25, true},
		{"25c", 25, true},
		{"25 derajat", 25, true},
		{"suhu 25", 25, true},
		{"temperatur 25", 25, true},
		{"dari 18 ke 25", 25, true}, // last number
		{"suhu 35", 35, true},       // even if out of range, we want to know user tried
		{"set suhu 25 selama 30 menit", 25, true},
		{"set AC 2 ke 25", 25, true},
		{"mode fan 3", 0, false},
		{"set kipas 3 suhu 24", 24, true},
		{"nyalakan", 0, false},
		{"mode 1", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.prompt, func(t *testing.T) {
			got, found := s.parseTemperature(tt.prompt)
			if found != tt.found {
				t.Errorf("parseTemperature() found = %v, want %v", found, tt.found)
			}
			if got != tt.want {
				t.Errorf("parseTemperature() got = %v, want %v", got, tt.want)
			}
		})
	}
}
