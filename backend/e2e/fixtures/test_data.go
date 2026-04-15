package fixtures

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

func RandomMAC() string {
	bytes := make([]byte, 6)
	rand.Read(bytes)
	return fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X", bytes[0], bytes[1], bytes[2], bytes[3], bytes[4], bytes[5])
}

func RandomRoomID() string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return "room-" + hex.EncodeToString(bytes)
}

func RandomUUID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

type TestTerminalFixture struct {
	MAC          string
	RoomID       string
	Name         string
	DeviceTypeID string
}

func NewTestTerminalFixture() *TestTerminalFixture {
	return &TestTerminalFixture{
		MAC:          RandomMAC(),
		RoomID:       RandomRoomID(),
		Name:         "Test Terminal " + time.Now().Format("15:04:05"),
		DeviceTypeID: "hub-type-001",
	}
}

func (f *TestTerminalFixture) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"mac_address":    f.MAC,
		"room_id":        f.RoomID,
		"name":           f.Name,
		"device_type_id": f.DeviceTypeID,
	}
}

type TestDeviceFixture struct {
	Name     string
	Category string
}

func NewTestDeviceFixture() *TestDeviceFixture {
	return &TestDeviceFixture{
		Name:     "E2E Test Device " + time.Now().Format("15:04:05"),
		Category: "switch",
	}
}

func (f *TestDeviceFixture) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"name":     f.Name,
		"category": f.Category,
	}
}
