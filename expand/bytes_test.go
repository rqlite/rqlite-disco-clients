package expand

import (
	"bytes"
	"os"
	"testing"
)

func Test_ExpandEnvBytes(t *testing.T) {
	testCases := []struct {
		name     string
		env      map[string]string
		input    []byte
		expected []byte
	}{
		{
			name: "Simple environment variables",
			env: map[string]string{
				"NAME": "John",
				"CITY": "San Francisco",
			},
			input:    []byte("Hello, $NAME! Welcome to $CITY."),
			expected: []byte("Hello, John! Welcome to San Francisco."),
		},
		{
			name:     "No environment variables",
			env:      map[string]string{},
			input:    []byte("Hello, World!"),
			expected: []byte("Hello, World!"),
		},
		{
			name: "Unknown environment variables",
			env: map[string]string{
				"NAME": "John",
			},
			input:    []byte("Greetings, $UNKNOWN!"),
			expected: []byte("Greetings, !"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set environment variables for each test case
			for key, value := range tc.env {
				os.Setenv(key, value)
			}

			output := ExpandEnvBytes(tc.input)
			if !bytes.Equal(output, tc.expected) {
				t.Errorf("Expected %q, but got %q", tc.expected, output)
			}

			// Clean up environment variables after each test case
			for key := range tc.env {
				os.Unsetenv(key)
			}
		})
	}
}
