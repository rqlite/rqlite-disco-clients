package consul

import (
	"strings"
	"testing"
)

func Test_LoadExampleConfig(t *testing.T) {
	r := strings.NewReader(exampleConfig)
	cfg, err := NewConfigFromReader(r)
	if err != nil {
		t.Fatalf("failed to generate config: %s", err.Error())
	}
	if cfg == nil {
		t.Fatalf("nil config")
	}
}