package utils

import "strings"

// MapCategoryToName converts a Tuya technical category code to a human-readable English name.
// This is used to improve similarity search accuracy in the Vector DB.
func MapCategoryToName(category string) string {
	category = strings.ToLower(category)
	switch category {
	case "infrared_ac":
		return "Air Conditioner"
	case "kg", "cz", "ws", "dj":
		return "Light / Switch / Socket"
	case "jsq":
		return "Humidifier"
	case "ms":
		return "Door Lock"
	case "sp":
		return "Camera"
	case "wnykq":
		return "Smart IR Hub"
	default:
		return "Smart Device"
	}
}
