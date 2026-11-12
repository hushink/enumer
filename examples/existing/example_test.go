// Copyright (c) 2026 The enumer Authors
// SPDX-License-Identifier: MIT

package existing

import "testing"

func TestExistingType(t *testing.T) {
	c, err := ParseColor("green")
	if err != nil || c != ColorGreen {
		t.Fatalf("parse: %v %v", c, err)
	}
	if ColorBlue.String() != "blue" {
		t.Fatalf("string: %q", ColorBlue.String())
	}
}
