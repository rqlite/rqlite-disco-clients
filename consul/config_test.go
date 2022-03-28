package consul

import (
	"strings"
	"testing"
)

const (
	badConfigHTTP = `
{
	"address": "http://1.2.3.4"
}
`
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
}

func Test_LoadBadConfigHTTP(t *testing.T) {
	r := strings.NewReader(badConfigHTTP)
	_, err := NewConfigFromReader(r)
	if err == nil {
		t.Fatalf("bad HTTP config unexpectedly parsed without error")
	}
}
