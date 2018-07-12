//go:generate go run maketables.go > tables.go

package confusables

import (
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// Skeleton converts a string to its skeleton form as described in
// http://www.unicode.org/reports/tr39/#Confusable_Detection
func Skeleton(s string) string {
	s = norm.NFKD.String(s)
	for i, w := 0, 0; i < len(s); i += w {
		char, width := utf8.DecodeRuneInString(s[i:])
		replacement, exists := confusablesMap[char]
		if exists {
			s = s[:i] + replacement + s[i+width:]
			w = len(replacement)
		} else {
			w = width
		}
	}
	s = norm.NFKD.String(s)

	return s
}
