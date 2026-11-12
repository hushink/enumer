// Copyright (c) 2026 The enumer Authors
// SPDX-License-Identifier: MIT

// Package existing demonstrates the default typeMode: auto. The Color type is
// declared by hand in types.go, so enumer generates only constants and methods.
// Remove types.go and enumer emits the type from the manifest instead.
//
//go:generate enumer generate
package existing
