package expand

import "os"

func ExpandEnvBytes(input []byte) []byte {
	// Convert byte slice to string
	inputStr := string(input)

	// Run os.ExpandEnv on the string
	expandedStr := os.ExpandEnv(inputStr)

	// Convert the expanded string back to a byte slice
	expandedBytes := []byte(expandedStr)

	return expandedBytes
}
