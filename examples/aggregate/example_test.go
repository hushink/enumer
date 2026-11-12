// Copyright (c) 2026 The enumer Authors
// SPDX-License-Identifier: MIT

package aggregate

import "testing"

func TestAggregate(t *testing.T) {
	if MethodGet.String() != "get" {
		t.Fatalf("method: %q", MethodGet.String())
	}
	if StatusOK.String() != "ok" {
		t.Fatalf("status: %q", StatusOK.String())
	}
	if StatusContinue != 100 || StatusAccepted != 103 {
		t.Fatalf("numeric values: %d %d", StatusContinue, StatusAccepted)
	}
	if !StatusOK.IsValid() || Status(999).IsValid() {
		t.Fatal("IsValid")
	}
}
