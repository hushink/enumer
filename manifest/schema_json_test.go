// Copyright (c) 2026 The enumer Authors
// SPDX-License-Identifier: MIT

package manifest

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/hushink/enumer/schema"
	"github.com/stretchr/testify/require"
)

func TestJSONSchemaDocument(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "schema", "enumer.schema.json"))
	require.NoError(t, err)

	var doc map[string]any
	require.NoError(t, json.Unmarshal(data, &doc))
	require.Equal(t, "http://json-schema.org/draft-07/schema#", doc["$schema"])
	require.Equal(t, schema.URL, doc["$id"])
	require.Equal(t, "object", doc["type"])
	require.Equal(t, schema.Bytes(), data)

	properties, ok := doc["properties"].(map[string]any)
	require.True(t, ok)
	require.Contains(t, properties, "$schema")
	require.Contains(t, properties, "enums")

	definitions, ok := doc["definitions"].(map[string]any)
	require.True(t, ok)
	require.Contains(t, definitions, "enum")
	require.Contains(t, definitions, "value")
}
