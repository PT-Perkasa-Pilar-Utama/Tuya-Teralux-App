package controllers

import (
	"sensio/domain/common/utils"
	"testing"
)

func TestResolveMQTTUID(t *testing.T) {
	orig := utils.AppConfig
	utils.AppConfig = &utils.Config{TuyaUserID: "sgFallback123456"}
	defer func() { utils.AppConfig = orig }()

	// Mock controller (terminalRepo nil is fine for these simple tests as it will hit fallback)
	c := &RAGChatController{}

	tests := []struct {
		name  string
		mac   string
		input string
		want  string
	}{
		{name: "valid tuya uid", mac: "", input: "sg1765086176746IkwBD", want: "sg1765086176746IkwBD"},
		{name: "hex terminal identity", mac: "", input: "3993F3DBEE8B8C74", want: "sgFallback123456"},
		{name: "mac format", mac: "", input: "AA:BB:CC:DD:EE:FF", want: "sgFallback123456"},
		{name: "empty", mac: "AA:BB:CC:DD:EE:FF", input: "", want: "sgFallback123456"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.resolveMQTTUID(tt.mac, tt.input)
			if got != tt.want {
				t.Fatalf("resolveMQTTUID(%q, %q) = %q, want %q", tt.mac, tt.input, got, tt.want)
			}
		})
	}
}
