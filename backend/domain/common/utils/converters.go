package utils

import (
	"strconv"
)

// ToInt converts an interface{} value to int.
// It supports int, int64, float64, and string types.
// Returns 0 and false if conversion fails.
func ToInt(v interface{}) (int, bool) {
	switch val := v.(type) {
	case int:
		return val, true
	case int64:
		return int(val), true
	case float64:
		return int(val), true
	case string:
		if i, err := strconv.Atoi(val); err == nil {
			return i, true
		}
	}
	return 0, false
}
