// Package diag defines structured diagnostics shared by the manifest loader,
// the generator and the command line interface.
package diag

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Severity classifies a diagnostic.
type Severity string

const (
	// Error means the operation cannot succeed.
	Error Severity = "error"
	// Warning means the operation can succeed but the input is suspicious.
	Warning Severity = "warning"
)

// Diagnostic is one structured report with an optional source location and a
// stable machine-readable code.
type Diagnostic struct {
	Severity Severity `json:"severity"`
	Code     string   `json:"code"`
	Message  string   `json:"message"`
	File     string   `json:"file,omitempty"`
	Line     int      `json:"line,omitempty"`
	Column   int      `json:"column,omitempty"`
	Enum     string   `json:"enum,omitempty"`
	Value    string   `json:"value,omitempty"`
}

// String renders a human-readable single-line diagnostic.
func (d Diagnostic) String() string {
	var b strings.Builder
	if d.File != "" {
		b.WriteString(d.File)
		if d.Line > 0 {
			fmt.Fprintf(&b, ":%d", d.Line)
			if d.Column > 0 {
				fmt.Fprintf(&b, ":%d", d.Column)
			}
		}
		b.WriteString(": ")
	}
	b.WriteString("enumer: ")
	if d.Severity == Warning {
		b.WriteString("warning: ")
	}
	b.WriteString(d.Message)
	return b.String()
}

// List is an ordered collection of diagnostics.
type List []Diagnostic

// HasErrors reports whether the list contains at least one error.
func (l List) HasErrors() bool {
	for _, d := range l {
		if d.Severity == Error {
			return true
		}
	}
	return false
}

// UpgradeWarnings converts every warning into an error. It is used by the
// --strict mode.
func (l List) UpgradeWarnings() List {
	out := make(List, len(l))
	for i, d := range l {
		if d.Severity == Warning {
			d.Severity = Error
		}
		out[i] = d
	}
	return out
}

// Sort orders diagnostics by file, line, column and code.
func (l List) Sort() {
	sort.SliceStable(l, func(i, j int) bool {
		a, b := l[i], l[j]
		if a.File != b.File {
			return a.File < b.File
		}
		if a.Line != b.Line {
			return a.Line < b.Line
		}
		if a.Column != b.Column {
			return a.Column < b.Column
		}
		if a.Code != b.Code {
			return a.Code < b.Code
		}
		return a.Message < b.Message
	})
}

// Format renders all diagnostics as newline-separated human-readable lines.
func (l List) Format() string {
	lines := make([]string, 0, len(l))
	for _, d := range l {
		lines = append(lines, d.String())
	}
	return strings.Join(lines, "\n")
}

// Report is the JSON envelope for diagnostics.
type Report struct {
	Version     int  `json:"version"`
	Diagnostics List `json:"diagnostics"`
}

// JSON renders the diagnostics as an indented JSON report.
func (l List) JSON() ([]byte, error) {
	if l == nil {
		l = List{}
	}
	return json.MarshalIndent(Report{Version: 1, Diagnostics: l}, "", "  ")
}
