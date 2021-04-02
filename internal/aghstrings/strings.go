// Package aghstrings contains utilities dealing with strings.
package aghstrings

import "strings"

// CloneSlice returns the copy of a.
func CloneSlice(a []string) []string {
	a2 := make([]string, len(a))
	copy(a2, a)
	return a2
}

// InSlice checks if string is in the slice of strings.
func InSlice(strs []string, str string) (ok bool) {
	for _, s := range strs {
		if s == str {
			return true
		}
	}

	return false
}

// SetSubtract subtracts b from a interpreted as sets.
func SetSubtract(a, b []string) (c []string) {
	// unit is an object to be used as value in set.
	type unit = struct{}

	cSet := make(map[string]unit)
	for _, k := range a {
		cSet[k] = unit{}
	}

	for _, k := range b {
		delete(cSet, k)
	}

	c = make([]string, len(cSet))
	i := 0
	for k := range cSet {
		c[i] = k
		i++
	}

	return c
}

// SplitNext splits string by a byte and returns the first chunk skipping empty
// ones.  Whitespaces are trimmed.
func SplitNext(str *string, splitBy byte) string {
	i := strings.IndexByte(*str, splitBy)
	s := ""
	if i != -1 {
		s = (*str)[0:i]
		*str = (*str)[i+1:]
		k := 0
		ch := rune(0)
		for k, ch = range *str {
			if byte(ch) != splitBy {
				break
			}
		}
		*str = (*str)[k:]
	} else {
		s = *str
		*str = ""
	}
	return strings.TrimSpace(s)
}
