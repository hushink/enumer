// Copyright (c) 2026 The enumer Authors
// SPDX-License-Identifier: MIT

package color

import "testing"

func TestColor(t *testing.T) {
	c, err := ParseColor("green")
	if err != nil || c != ColorGreen {
		t.Fatalf("parse: %v %v", c, err)
	}
}
