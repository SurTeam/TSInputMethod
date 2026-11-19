package emoji

import (
	"bufio"
	"os"
	"strings"
)

type Table struct {
	byText map[string][]string
}

func LoadFile(path string) (*Table, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	byText := make(map[string][]string)
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "----------") {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		if key == "" {
			continue
		}
		emojis := make([]string, 0, 4)
		seen := map[string]struct{}{}
		for _, token := range strings.Fields(parts[1]) {
			if token == key {
				continue
			}
			if _, ok := seen[token]; ok {
				continue
			}
			seen[token] = struct{}{}
			emojis = append(emojis, token)
		}
		if len(emojis) == 0 {
			continue
		}
		prev := byText[key]
		for _, emoji := range emojis {
			dup := false
			for _, existing := range prev {
				if existing == emoji {
					dup = true
					break
				}
			}
			if !dup {
				prev = append(prev, emoji)
			}
		}
		byText[key] = prev
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return &Table{byText: byText}, nil
}

func (t *Table) Lookup(text string) []string {
	if t == nil || text == "" {
		return nil
	}
	out := t.byText[text]
	if len(out) == 0 {
		return nil
	}
	return append([]string(nil), out...)
}

func (t *Table) Size() int {
	if t == nil {
		return 0
	}
	return len(t.byText)
}
