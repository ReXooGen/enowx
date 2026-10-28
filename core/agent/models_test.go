package agent

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestOMPModelsOnlyExposeRegisteredChatMetadata(t *testing.T) {
	catalog := []byte(`{"models":[
{"provider":"custom","id":"model-a","name":"Model A","selector":"custom/model-a","kind":"chat","apiKey":"secret-do-not-send","baseUrl":"https://private.example"},
{"provider":"custom","id":"model-a","selector":"custom/model-a"},
{"provider":"other","id":"model-b","selector":"other/model-b","kind":"chat"},
{"provider":"other","id":"embed","kind":"embedding"},
{"provider":"other","id":"wrong","selector":"injected/different"},
{"provider":"bad\nprovider","id":"invalid"}
]}`)
	got, err := parseOMPModels(catalog)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Registered || len(got.Models) != 2 {
		t.Fatalf("registered models: %+v", got)
	}
	if got.Models[0].Selector != "custom/model-a" || got.Models[1].Name != "model-b" {
		t.Fatalf("identifiers changed: %+v", got)
	}
	data, _ := json.Marshal(got)
	if strings.Contains(string(data), "secret-do-not-send") || strings.Contains(string(data), "private.example") {
		t.Fatal("catalog leaked provider configuration")
	}
	if _, err = parseOMPModels([]byte("CLI startup failed")); err == nil {
		t.Fatal("unreadable CLI output accepted")
	}
}
