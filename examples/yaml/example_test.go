// Copyright (c) 2026 The enumer Authors
// SPDX-License-Identifier: MIT

package yamlenum

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestYAML(t *testing.T) {
	data, err := yaml.Marshal(ColorGreen)
	if err != nil || string(data) != "green\n" {
		t.Fatalf("marshal: %q %v", data, err)
	}

	var c Color
	if err := yaml.Unmarshal([]byte("green"), &c); err != nil || c != ColorGreen {
		t.Fatalf("unmarshal: %v %v", c, err)
	}
	if err := yaml.Unmarshal([]byte("nope"), &c); err == nil {
		t.Fatal("expected invalid value error")
	}
}
