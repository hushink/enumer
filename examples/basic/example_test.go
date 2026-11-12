// Copyright (c) 2026 The enumer Authors
// SPDX-License-Identifier: MIT

package basic

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestColor(t *testing.T) {
	c, err := ParseColor("g")
	if err != nil || c != ColorGreen || c.String() != "GREEN" {
		t.Fatalf("parse: %v %v", c, err)
	}
	if _, err := ParseColor("nope"); !errors.Is(err, ErrInvalidColor) {
		t.Fatalf("expected ErrInvalidColor, got %v", err)
	}
	data, err := json.Marshal(ColorGreen)
	if err != nil || string(data) != `"GREEN"` {
		t.Fatalf("json: %s %v", data, err)
	}
	var decoded Color
	if err := json.Unmarshal([]byte(`"red"`), &decoded); err != nil || decoded != ColorRed {
		t.Fatalf("unmarshal: %v %v", decoded, err)
	}
}

func TestStatus(t *testing.T) {
	if !StatusFailed.IsValid() || Status(5).IsValid() {
		t.Fatal("IsValid")
	}
	data, err := json.Marshal(StatusFailed)
	if err != nil || string(data) != "10" {
		t.Fatalf("json: %s %v", data, err)
	}
	var n NullStatus
	if err := n.Scan(int64(10)); err != nil || !n.Valid || n.Enum != StatusFailed {
		t.Fatalf("scan: %v %v %v", n.Enum, n.Valid, err)
	}
	if v, err := n.Value(); err != nil || v.(int64) != 10 {
		t.Fatalf("value: %v %v", v, err)
	}
}
