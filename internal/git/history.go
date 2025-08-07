package git

import (
	"fmt"
	"os/exec"
	"strings"
)

type Entry struct {
	Hash  string
	Files map[string][]byte
}

func LastNCommits(root string, n int) ([]Entry, error) {
	if n <= 0 {
		return nil, nil
	}
	// Use `git show` per commit to keep it simple
	cmd := exec.Command("git", "-C", root, "rev-list", "--max-count", fmt.Sprintf("%d", n), "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	hashes := strings.Fields(string(out))

	var entries []Entry
	for _, h := range hashes {
		// get changed files + content in commit
		cmd = exec.Command("git", "-C", root, "show", h, "--name-only", "--pretty=")
		filesOut, err := cmd.Output()
		if err != nil {
			continue
		}
		fileList := strings.Fields(string(filesOut))
		files := map[string][]byte{}
		for _, p := range fileList {
			show := exec.Command("git", "-C", root, "show", h+":"+p)
			b, err := show.Output()
			if err == nil {
				files[p] = b
			}
		}
		entries = append(entries, Entry{Hash: h, Files: files})
	}
	return entries, nil
}

func DiffAgainst(root, base string) ([]string, [][]byte, error) {
	cmd := exec.Command("git", "-C", root, "diff", "--name-only", base)
	out, err := cmd.Output()
	if err != nil {
		return nil, nil, err
	}
	paths := strings.Fields(string(out))
	var data [][]byte
	for _, p := range paths {
		show := exec.Command("git", "-C", root, "diff", base, "--", p)
		b, err := show.Output()
		if err != nil {
			b = []byte{}
		}
		// crude: take added lines only could be a later enhancement
		data = append(data, b)
	}
	return paths, data, nil
}

func StagedDiff(root string) ([]string, [][]byte, error) {
	cmd := exec.Command("git", "-C", root, "diff", "--name-only", "--cached")
	out, err := cmd.Output()
	if err != nil {
		return nil, nil, err
	}
	paths := strings.Fields(string(out))
	var data [][]byte
	for _, p := range paths {
		show := exec.Command("git", "-C", root, "show", ":"+p)
		b, err := show.Output()
		if err != nil {
			b = []byte{}
		}
		data = append(data, b)
	}
	return paths, data, nil
}
