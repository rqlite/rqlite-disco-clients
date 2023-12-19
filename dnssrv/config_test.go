package dnssrv

import (
	"strings"
	"testing"
)

func Test_NilReaderConfig(t *testing.T) {
	cfg, err := NewConfigFromReader(nil)
	if err != nil {
		t.Fatalf("failed to generate config: %s", err.Error())
	}
	if cfg != nil {
		t.Fatalf("expected nil config")
	}
}

func Test_LoadExampleConfig(t *testing.T) {
	r := strings.NewReader(exampleConfig)
	cfg, err := NewConfigFromReader(r)
	if err != nil {
		t.Fatalf("failed to generate config: %s", err.Error())
	}
	if cfg == nil {
		t.Fatalf("nil config")
	}

	if cfg.Name != "rqlite.com" || cfg.Service != "rqlite-raft" {
		t.Fatalf("invalid config generated")
	}
}
