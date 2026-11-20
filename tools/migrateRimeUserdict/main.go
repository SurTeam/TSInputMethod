package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/SurTeam/TSInputMethod/internal/dictionary"
	"github.com/SurTeam/TSInputMethod/internal/pinyin"
)

const (
	formatVersion = 1
	defaultWeight = uint32(1000)
	maxTextRunes  = 16
	maxSyllables  = 12
	maxUsage      = uint32(100000)
	sourceLearned = "learned"
	sourceManual  = "manual"
)

type fileDoc struct {
	FormatVersion int         `json:"format_version"`
	UpdatedAt     time.Time   `json:"updated_at"`
	Entries       []fileEntry `json:"entries"`
}

type fileEntry struct {
	Text       string    `json:"text"`
	Syllables  []string  `json:"syllables"`
	Weight     uint32    `json:"weight"`
	UsageCount uint32    `json:"usage_count"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Source     string    `json:"source"`
	Enabled    bool      `json:"enabled"`
}

type candidate struct {
	Text      string
	Syllables []string
	Usage     uint32
}

var usageRe = regexp.MustCompile(`c=(\d+)`)

func main() {
	dictionary.SetValidSyllableFunc(pinyin.IsValidSyllable)
	source := ""
	dest := ""
	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--source":
			if i+1 >= len(args) {
				fatalf("--source needs a path")
			}
			source = args[i+1]
			i++
		case "--dest":
			if i+1 >= len(args) {
				fatalf("--dest needs a path")
			}
			dest = args[i+1]
			i++
		default:
			fatalf("unknown arg %q", args[i])
		}
	}
	if source == "" {
		fatalf("missing --source")
	}
	if dest == "" {
		var err error
		dest, err = dictionary.DefaultUserDictPath()
		if err != nil {
			fatalf("default dest: %v", err)
		}
	}
	candidates, skipped, err := parseRimeUserDB(source)
	if err != nil {
		fatalf("parse source: %v", err)
	}
	existing, err := loadExisting(dest)
	if err != nil {
		fatalf("load dest: %v", err)
	}
	merged, stats := mergeEntries(existing, candidates)
	if err := writeAtomic(dest, merged); err != nil {
		fatalf("write dest: %v", err)
	}
	fmt.Printf("source=%s\n", source)
	fmt.Printf("dest=%s\n", dest)
	fmt.Printf("parsed_ok=%d skipped_invalid=%d\n", len(candidates), skipped)
	fmt.Printf("existing=%d added=%d usage_updated=%d truncated=%d final=%d\n",
		stats.existing, stats.added, stats.usageUpdated, stats.truncated, len(merged))
}

type mergeStats struct {
	existing     int
	added        int
	usageUpdated int
	truncated    int
}

func mergeEntries(existing []fileEntry, candidates []candidate) ([]fileEntry, mergeStats) {
	now := time.Now().UTC()
	byKey := make(map[string]int, len(existing)+len(candidates))
	out := make([]fileEntry, 0, len(existing)+len(candidates))
	stats := mergeStats{existing: len(existing)}
	for _, e := range existing {
		key := entryKey(e.Text, e.Syllables)
		if _, ok := byKey[key]; ok {
			continue
		}
		byKey[key] = len(out)
		out = append(out, e)
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Usage == candidates[j].Usage {
			if candidates[i].Text == candidates[j].Text {
				return strings.Join(candidates[i].Syllables, "'") < strings.Join(candidates[j].Syllables, "'")
			}
			return candidates[i].Text < candidates[j].Text
		}
		return candidates[i].Usage > candidates[j].Usage
	})
	for _, c := range candidates {
		key := entryKey(c.Text, c.Syllables)
		if idx, ok := byKey[key]; ok {
			if c.Usage > out[idx].UsageCount {
				out[idx].UsageCount = c.Usage
				out[idx].UpdatedAt = now
				stats.usageUpdated++
			}
			continue
		}
		out = append(out, fileEntry{
			Text:       c.Text,
			Syllables:  append([]string(nil), c.Syllables...),
			Weight:     defaultWeight,
			UsageCount: c.Usage,
			CreatedAt:  now,
			UpdatedAt:  now,
			Source:     sourceLearned,
			Enabled:    true,
		})
		byKey[key] = len(out) - 1
		stats.added++
	}
	return out, stats
}

func parseRimeUserDB(path string) ([]candidate, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}
	defer f.Close()
	best := make(map[string]candidate)
	skipped := 0
	scanner := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 3 {
			skipped++
			continue
		}
		syllables := strings.Fields(strings.TrimSpace(parts[0]))
		text := strings.TrimSpace(parts[1])
		usage := uint32(0)
		if m := usageRe.FindStringSubmatch(parts[2]); len(m) == 2 {
			n, err := strconv.ParseUint(m[1], 10, 32)
			if err == nil {
				usage = uint32(n)
			}
		}
		if usage > maxUsage {
			usage = maxUsage
		}
		text, err := normalizeText(text)
		if err != nil {
			skipped++
			continue
		}
		syllables, err = normalizeSyllables(syllables)
		if err != nil {
			skipped++
			continue
		}
		key := entryKey(text, syllables)
		if prev, ok := best[key]; ok {
			if usage > prev.Usage {
				prev.Usage = usage
				best[key] = prev
			}
			continue
		}
		best[key] = candidate{Text: text, Syllables: syllables, Usage: usage}
	}
	if err := scanner.Err(); err != nil {
		return nil, 0, err
	}
	out := make([]candidate, 0, len(best))
	for _, c := range best {
		out = append(out, c)
	}
	return out, skipped, nil
}

func loadExisting(path string) ([]fileEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	var doc fileDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	out := make([]fileEntry, 0, len(doc.Entries))
	for _, e := range doc.Entries {
		text, err := normalizeText(e.Text)
		if err != nil {
			continue
		}
		syllables, err := normalizeSyllables(e.Syllables)
		if err != nil {
			continue
		}
		source := e.Source
		if source != sourceManual && source != sourceLearned {
			source = sourceManual
		}
		weight := e.Weight
		if weight == 0 {
			weight = defaultWeight
		}
		out = append(out, fileEntry{
			Text:       text,
			Syllables:  syllables,
			Weight:     weight,
			UsageCount: e.UsageCount,
			CreatedAt:  e.CreatedAt.UTC(),
			UpdatedAt:  e.UpdatedAt.UTC(),
			Source:     source,
			Enabled:    e.Enabled || e.CreatedAt.IsZero(),
		})
	}
	return out, nil
}

func writeAtomic(path string, entries []fileEntry) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	doc := fileDoc{
		FormatVersion: formatVersion,
		UpdatedAt:     time.Now().UTC(),
		Entries:       entries,
	}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp, err := os.CreateTemp(filepath.Dir(path), "user_dictionary-migrate-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

func normalizeText(text string) (string, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return "", fmt.Errorf("empty text")
	}
	runess := []rune(text)
	if len(runess) == 0 || len(runess) > maxTextRunes {
		return "", fmt.Errorf("text length")
	}
	for _, r := range runess {
		if unicode.IsControl(r) {
			return "", fmt.Errorf("control rune")
		}
	}
	return text, nil
}

func normalizeSyllables(syllables []string) ([]string, error) {
	out := make([]string, 0, len(syllables))
	for _, syl := range syllables {
		syl = strings.ToLower(strings.TrimSpace(syl))
		if syl == "" {
			return nil, fmt.Errorf("empty syllable")
		}
		out = append(out, syl)
	}
	if err := dictionary.ValidateUserSyllables(out); err != nil {
		return nil, err
	}
	return out, nil
}

func entryKey(text string, syllables []string) string {
	return text + "\x00" + strings.Join(syllables, "'")
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "migrateRimeUserdict: "+format+"\n", args...)
	os.Exit(1)
}
