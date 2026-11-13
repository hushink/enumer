// Copyright (c) 2026 The enumer Authors
// SPDX-License-Identifier: MIT

// Package schema exposes the JSON Schema for enumer.yaml manifests.
package schema

import _ "embed"

// URL is the canonical published URL of the manifest JSON Schema.
const URL = "https://raw.githubusercontent.com/hushink/enumer/main/schema/enumer.schema.json"

// FileName is the schema file name written by enumer init.
const FileName = "enumer.schema.json"

// data is the embedded copy of enumer.schema.json.
//
//go:embed enumer.schema.json
var data []byte

// Bytes returns a copy of the embedded JSON Schema.
func Bytes() []byte {
	return append([]byte(nil), data...)
}
