// Package buildinfo carries version metadata injected at build time.
package buildinfo

import "fmt"

// These values are overridden with -ldflags at release time.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// String returns the full version string.
func String() string {
	if Commit == "" || Commit == "none" {
		return Version
	}
	return fmt.Sprintf("%s (commit %s, built %s)", Version, Commit, Date)
}
