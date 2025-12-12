package utils

import "strings"

// JoinStrings joins a slice of strings with a separator
func JoinStrings(elems []string, sep string) string {
	return strings.Join(elems, sep)
}
