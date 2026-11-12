// Package naming provides identifier conversions used by the manifest loader,
// the generator and custom template functions.
package naming

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// DefaultInitialisms is the set of common initialisms that are upper-cased as
// a whole when converting value names to Go identifiers.
//
// The list follows the well-known golint initialisms and can be extended with
// the manifest option naming.extraInitialisms.
var DefaultInitialisms = []string{
	"ACL", "API", "ASCII", "CPU", "CSS", "DNS", "EOF", "GUID", "HTML", "HTTP",
	"HTTPS", "ID", "IP", "JSON", "LHS", "QPS", "RAM", "RHS", "RPC", "SLA",
	"SMTP", "SQL", "SSH", "TCP", "TLS", "TTL", "UDP", "UI", "UID", "UUID",
	"URI", "URL", "UTF8", "VM", "XML", "XMPP", "XSRF", "XSS",
}

// InitialismSet builds a lookup set from the default initialisms plus extras.
// When disabled is true the set is empty regardless of extras.
func InitialismSet(extra []string, disabled bool) map[string]bool {
	if disabled {
		return map[string]bool{}
	}
	set := make(map[string]bool, len(DefaultInitialisms)+len(extra))
	for _, s := range DefaultInitialisms {
		set[s] = true
	}
	for _, s := range extra {
		s = strings.ToUpper(strings.TrimSpace(s))
		if s != "" {
			set[s] = true
		}
	}
	return set
}

// SplitWords splits a name on separators and camel-case boundaries.
//
// Examples:
//
//	"http_status"  -> ["http", "status"]
//	"HTTPStatus"   -> ["HTTP", "Status"]
//	"redOrange"    -> ["red", "Orange"]
//	"in-progress"  -> ["in", "progress"]
func SplitWords(s string) []string {
	runes := []rune(s)
	words := make([]string, 0, 4)
	current := make([]rune, 0, len(runes))
	flush := func() {
		if len(current) > 0 {
			words = append(words, string(current))
			current = current[:0]
		}
	}
	for i, r := range runes {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			flush()
			continue
		}
		if len(current) > 0 {
			prev := current[len(current)-1]
			var next rune
			if i+1 < len(runes) {
				next = runes[i+1]
			}
			if unicode.IsUpper(r) {
				if unicode.IsLower(prev) || unicode.IsDigit(prev) ||
					(unicode.IsUpper(prev) && unicode.IsLower(next)) {
					flush()
				}
			}
		}
		current = append(current, r)
	}
	flush()
	return words
}

// Camel converts s to CamelCase, upper-casing known initialisms.
func Camel(s string, initialisms map[string]bool) string {
	var b strings.Builder
	for _, word := range SplitWords(s) {
		upper := strings.ToUpper(word)
		if initialisms[upper] {
			b.WriteString(upper)
			continue
		}
		b.WriteString(upperFirst(strings.ToLower(word)))
	}
	return b.String()
}

// Snake converts s to snake_case.
func Snake(s string) string {
	return strings.ToLower(strings.Join(SplitWords(s), "_"))
}

// Kebab converts s to kebab-case.
func Kebab(s string) string {
	return strings.ToLower(strings.Join(SplitWords(s), "-"))
}

// Title converts s to space-separated Title Case.
func Title(s string) string {
	words := SplitWords(s)
	for i, word := range words {
		words[i] = upperFirst(strings.ToLower(word))
	}
	return strings.Join(words, " ")
}

func upperFirst(s string) string {
	if s == "" {
		return ""
	}
	r, size := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[size:]
}
