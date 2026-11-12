// Copyright (c) 2026 The enumer Authors
// SPDX-License-Identifier: MIT

package flags

import (
	"encoding/json"
	"testing"
)

func TestPermission(t *testing.T) {
	rw := PermissionRead | PermissionWrite
	if rw.String() != "read|write" {
		t.Fatalf("string: %q", rw.String())
	}

	parsed, err := ParsePermission("read|execute")
	if err != nil || parsed != PermissionRead|PermissionExecute {
		t.Fatalf("parse: %v %v", parsed, err)
	}
	if !rw.IsValid() || Permission(8).IsValid() {
		t.Fatal("IsValid")
	}

	data, err := json.Marshal(PermissionNone)
	if err != nil || string(data) != "0" {
		t.Fatalf("json: %s %v", data, err)
	}
}
