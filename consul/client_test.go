package consul

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"
)

func Test_NewClient(t *testing.T) {
	c, err := New("rqlite", nil)
	if err != nil {
		t.Fatalf("failed to create new client: %s", err.Error())
	}
	if c == nil {
		t.Fatalf("returned client is nil")
	}
	if got, exp := c.String(), "consul-kv"; got != exp {
		t.Fatalf("wrong name for client, got %s, exp %s", got, exp)
	}
	if err := c.Close(); err != nil {
		t.Fatalf("failed to close client: %s", err.Error())
	}
}

func Test_NewClientConfigConnectOK(t *testing.T) {
	cfgFile := mustWriteConfigToTmpFile(&Config{
		Address: "localhost:8500",
		Scheme:  "http",
	})
	defer os.Remove(cfgFile)

	cfg, err := NewConfigFromFile(cfgFile)
	if err != nil {
		t.Fatalf("failed to get config from file: %s", err.Error())
	}

	client, err := New(randomString(), cfg)
	if err != nil {
		t.Fatalf("failed to create new client with config: %s", err.Error())
	}

	err = client.SetLeader("2", "http://localhost:4003", "localhost:4004")
	if err != nil {
		t.Fatalf("error when setting leader: %s", err.Error())
	}
}

func Test_NewClientConfigConnectOKEnv(t *testing.T) {
	t.Setenv("CONSUL_ADDRESS", "localhost:8500")
	cfgFile := mustWriteConfigToTmpFile(&Config{
		Address: "${CONSUL_ADDRESS}",
		Scheme:  "http",
	})
	defer os.Remove(cfgFile)

	cfg, err := NewConfigFromFile(cfgFile)
	if err != nil {
		t.Fatalf("failed to get config from file: %s", err.Error())
	}

	client, err := New(randomString(), cfg)
	if err != nil {
		t.Fatalf("failed to create new client with config: %s", err.Error())
	}

	err = client.SetLeader("2", "http://localhost:4003", "localhost:4004")
	if err != nil {
		t.Fatalf("error when setting leader: %s", err.Error())
	}
}

func Test_NewClientConfigReaderConnectOK(t *testing.T) {
	cfgFile := mustWriteConfigToTmpFile(&Config{
		Address: "localhost:8500",
		Scheme:  "http",
	})
	defer os.Remove(cfgFile)

	reader, err := os.Open(cfgFile)
	if err != nil {
		t.Fatalf("failed to open config file: %s", err.Error())
	}
	defer reader.Close()

	cfg, err := NewConfigFromReader(reader)
	if err != nil {
		t.Fatalf("failed to get config from file: %s", err.Error())
	}

	client, err := New(randomString(), cfg)
	if err != nil {
		t.Fatalf("failed to create new client with config: %s", err.Error())
	}

	err = client.SetLeader("2", "http://localhost:4003", "localhost:4004")
	if err != nil {
		t.Fatalf("error when setting leader: %s", err.Error())
	}
}

func Test_NewClientConfigConnectFail(t *testing.T) {
	cfgFile := mustWriteConfigToTmpFile(&Config{
		Address: "localhost:8501",
		Scheme:  "http",
	})
	defer os.Remove(cfgFile)

	cfg, err := NewConfigFromFile(cfgFile)
	if err != nil {
		t.Fatalf("failed to get config from file: %s", err.Error())
	}

	client, err := New(randomString(), cfg)
	if err != nil {
		t.Fatalf("failed to create new client with config: %s", err.Error())
	}

	err = client.SetLeader("2", "http://localhost:4003", "localhost:4004")
	if err == nil {
		t.Fatalf("should have failed to connect to consul")
	}
}

func Test_InitializeLeader(t *testing.T) {
	c, _ := New(randomString(), nil)
	defer c.Close()
	_, _, _, ok, err := c.GetLeader()
	if err != nil {
		t.Fatalf("failed to GetLeader: %s", err.Error())
	}
	if ok {
		t.Fatalf("leader found when not expected")
	}

	ok, err = c.InitializeLeader("1", "http://localhost:4001", "localhost:4002")
	if err != nil {
		t.Fatalf("error when initializing leader: %s", err.Error())
	}
	if !ok {
		t.Fatalf("failed to initialize leader")
	}

	id, api, addr, ok, err := c.GetLeader()
	if err != nil {
		t.Fatalf("failed to GetLeader: %s", err.Error())
	}
	if !ok {
		t.Fatalf("leader not found when expected")
	}
	if id != "1" || api != "http://localhost:4001" || addr != "localhost:4002" {
		t.Fatalf("retrieved incorrect details for leader")
	}
}

func Test_InitializeLeaderConflict(t *testing.T) {
	c, _ := New(randomString(), nil)
	defer c.Close()
	_, _, _, ok, err := c.GetLeader()
	if err != nil {
		t.Fatalf("failed to GetLeader: %s", err.Error())
	}
	if ok {
		t.Fatalf("leader found when not expected")
	}

	err = c.SetLeader("2", "http://localhost:4003", "localhost:4004")
	if err != nil {
		t.Fatalf("error when setting leader: %s", err.Error())
	}

	ok, err = c.InitializeLeader("1", "http://localhost:4001", "localhost:4002")
	if err != nil {
		t.Fatalf("error when initializing leader: %s", err.Error())
	}
	if ok {
		t.Fatalf("initialized leader when should have failed")
	}

	id, api, addr, ok, err := c.GetLeader()
	if err != nil {
		t.Fatalf("failed to GetLeader: %s", err.Error())
	}
	if !ok {
		t.Fatalf("leader not found when expected")
	}
	if id != "2" || api != "http://localhost:4003" || addr != "localhost:4004" {
		t.Fatalf("retrieved incorrect details for leader")
	}
}

func randomString() string {
	rand.Seed(time.Now().UnixNano())
	var output strings.Builder
	chars := "abcdedfghijklmnopqrstABCDEFGHIJKLMNOP"
	for i := 0; i < 10; i++ {
		random := rand.Intn(len(chars))
		randomChar := chars[random]
		output.WriteString(string(randomChar))
	}
	return output.String()
}

func mustWriteConfigToTmpFile(cfg *Config) string {
	f := mustTempFile()
	b, err := json.MarshalIndent(cfg, "", " ")
	if err != nil {
		panic("failed to JSON marshal config")
	}
	if err := ioutil.WriteFile(f, b, 0644); err != nil {
		panic("failed to write JSON to file")
	}
	return f
}

func mustTempFile() string {
	tmpfile, err := ioutil.TempFile("", "rqlite-db-test")
	if err != nil {
		panic(err.Error())
	}
	tmpfile.Close()
	return tmpfile.Name()
}
