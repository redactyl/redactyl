package redact

import (
	"os"
	"regexp"
)

type Replacement struct {
	Pattern *regexp.Regexp
	Replace string
}

// Apply applies replacements to a file on disk. Returns true if changed.
func Apply(path string, reps []Replacement) (bool, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	orig := string(b)
	s := orig
	for _, r := range reps {
		s = r.Pattern.ReplaceAllString(s, r.Replace)
	}
	if s == orig {
		return false, nil
	}
	return true, os.WriteFile(path, []byte(s), 0644)
}

// WouldChange returns true if applying replacements would modify the file, without writing.
func WouldChange(path string, reps []Replacement) (bool, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	orig := string(b)
	s := orig
	for _, r := range reps {
		s = r.Pattern.ReplaceAllString(s, r.Replace)
	}
	return s != orig, nil
}
