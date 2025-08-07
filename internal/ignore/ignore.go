package ignore

import (
	"os"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

type Matcher struct{ ps []gitignore.Pattern }

func Load(path string) (Matcher, error) {
	var m Matcher
	// load .redactylignore if present; fallback to none
	// simple: use gitignore parser; non-existent is fine
	data, err := os.ReadFile(path)
	if err != nil {
		return m, nil
	}
	var ps []gitignore.Pattern
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		ps = append(ps, gitignore.ParsePattern(line, nil))
	}
	m.ps = ps
	return m, nil
}

func (m Matcher) Match(p string) bool {
	for _, pat := range m.ps {
		if pat.Match(strings.Split(p, "/"), false) {
			return true
		}
	}
	return false
}
