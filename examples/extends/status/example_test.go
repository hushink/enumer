package status

import (
	"encoding/json"
	"testing"
)

func TestStatus(t *testing.T) {
	if !StatusDone.IsValid() {
		t.Fatal("IsValid")
	}
	data, err := json.Marshal(StatusDone)
	if err != nil || string(data) != "2" {
		t.Fatalf("json: %s %v", data, err)
	}
}
