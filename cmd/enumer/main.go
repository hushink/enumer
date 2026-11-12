// Copyright (c) 2026 The enumer Authors
// SPDX-License-Identifier: MIT

package main

import (
	"os"

	"github.com/hushink/enumer/internal/cli"
)

func main() {
	os.Exit(cli.Execute(os.Args[1:], os.Stdout, os.Stderr))
}
