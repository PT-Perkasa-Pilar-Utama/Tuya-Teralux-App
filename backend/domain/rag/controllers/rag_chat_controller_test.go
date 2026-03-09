package controllers

import (
	"sensio/domain/common/utils"
	"testing"
)

func TestResolveMQTTUID(t *testing.T) {
	orig := utils.AppConfig
	utils.AppConfig = &utils.Config{TuyaUserID: "sgFallback123456"}
	defer func() { utils.AppConfig = orig }()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "valid tuya uid", input: "sg1765086176746IkwBD", want: "sg1765086176746IkwBD"},
		{name: "hex terminal identity", input: "3993F3DBEE8B8C74", want: "sgFallback123456"},
		{name: "mac format", input: "AA:BB:CC:DD:EE:FF", want: "sgFallback123456"},
		{name: "empty", input: "", want: "sgFallback123456"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveMQTTUID(tt.input)
			if got != tt.want {
				t.Fatalf("resolveMQTTUID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
