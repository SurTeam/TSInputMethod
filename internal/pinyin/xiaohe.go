package pinyin

import "strings"

var xiaoheInitials = map[byte]string{
	'b': "b", 'p': "p", 'm': "m", 'f': "f",
	'd': "d", 't': "t", 'n': "n", 'l': "l",
	'g': "g", 'k': "k", 'h': "h",
	'j': "j", 'q': "q", 'x': "x",
	'r': "r", 'z': "z", 'c': "c", 's': "s",
	'y': "y", 'w': "w",
	'v': "zh", 'i': "ch", 'u': "sh",
}

var xiaoheFinals = map[byte]string{
	'a': "a", 'o': "o", 'e': "e", 'i': "i", 'u': "u", 'v': "v",
	'q': "iu", 'w': "ei", 'r': "uan", 't': "ve", 'y': "un",
	'p': "ie", 's': "ong", 'd': "ai", 'f': "en", 'g': "eng",
	'h': "ang", 'j': "an", 'k': "ing", 'l': "uang",
	'z': "ou", 'x': "ia", 'c': "ao", 'b': "in", 'n': "iao", 'm': "ian",
}

var xiaoheZeroSyllables = map[string]string{
		"aa": "a", "oo": "o", "ee": "e",
		"ah": "ang", "ag": "eng", "as": "ong",
		"aj": "an", "ac": "ao", "ad": "ai",
		"ew": "ei", "ef": "en", "eg": "eng", "er": "er", "ez": "ou",
		"oz": "ou",
		"ai": "ai", "ei": "ei", "ao": "ao", "ou": "ou",
		"an": "an", "en": "en",
	}

func ExpandXiaohe(input string) (syllables []string, remainder string, ok bool) {
	paths, rem, ok := ExpandXiaohePaths(input)
	if !ok || len(paths) == 0 {
		return nil, rem, ok
	}
	return append([]string(nil), paths[0]...), rem, true
}

func ExpandXiaohePaths(input string) (paths [][]string, remainder string, ok bool) {
	input = strings.ToLower(strings.TrimSpace(input))
	if input == "" {
		return [][]string{nil}, "", true
	}
	for _, r := range input {
		if r < 'a' || r > 'z' {
			return nil, input, false
		}
	}
	paths = [][]string{nil}
	i := 0
	for i+1 < len(input) {
		pair := input[i : i+2]
		alts, mapped := mapXiaohePairAll(pair[0], pair[1])
		if !mapped || len(alts) == 0 {
			if len(paths) == 1 && len(paths[0]) == 0 {
				return nil, input[i:], false
			}
			return paths, input[i:], false
		}
		next := make([][]string, 0, len(paths)*len(alts))
		for _, prefix := range paths {
			for _, syl := range alts {
				item := append(append([]string(nil), prefix...), syl)
				next = append(next, item)
				if len(next) >= 32 {
					break
				}
			}
			if len(next) >= 32 {
				break
			}
		}
		paths = next
		i += 2
	}
	if i < len(input) {
		return paths, mapXiaoheRemainder(input[i:]), true
	}
	return paths, "", true
}

func mapXiaoheRemainder(rem string) string {
	if len(rem) != 1 {
		return rem
	}
	if init, ok := xiaoheInitials[rem[0]]; ok {
		return init
	}
	return rem
}

func mapXiaohePair(first, second byte) (string, bool) {
	alts, ok := mapXiaohePairAll(first, second)
	if !ok || len(alts) == 0 {
		return "", false
	}
	return alts[0], true
}

func mapXiaohePairAll(first, second byte) ([]string, bool) {
	if syl, ok := xiaoheZeroSyllables[string([]byte{first, second})]; ok {
		return []string{syl}, true
	}
	init, okInit := xiaoheInitials[first]
	if !okInit {
		return nil, false
	}
	fins, okFin := resolveXiaoheFinals(init, second)
	if !okFin || len(fins) == 0 {
		return nil, false
	}
	out := make([]string, 0, len(fins))
	seen := map[string]struct{}{}
	for _, fin := range fins {
		syl := fin
		if init != "" {
			syl = init + fin
		}
		if !IsValidSyllable(syl) {
			continue
		}
		if _, dup := seen[syl]; dup {
			continue
		}
		seen[syl] = struct{}{}
		out = append(out, syl)
	}
	if len(out) == 0 {
		return nil, false
	}
	return out, true
}

func resolveXiaoheFinal(initial string, key byte) (string, bool) {
	fins, ok := resolveXiaoheFinals(initial, key)
	if !ok || len(fins) == 0 {
		return "", false
	}
	return fins[0], true
}

func resolveXiaoheFinals(initial string, key byte) ([]string, bool) {
	switch key {
	case 'o':
		if initial == "" || initial == "b" || initial == "p" || initial == "m" || initial == "f" || initial == "w" || initial == "y" {
			return preferredFinals(initial, "o", "uo"), true
		}
		return preferredFinals(initial, "uo", "o"), true
	case 's':
		if initial == "j" || initial == "q" || initial == "x" {
			return preferredFinals(initial, "iong", "ong"), true
		}
		return preferredFinals(initial, "ong", "iong"), true
	case 'k':
		if initial == "g" || initial == "k" || initial == "h" || initial == "zh" || initial == "ch" || initial == "sh" || initial == "r" || initial == "z" || initial == "c" || initial == "s" {
			return preferredFinals(initial, "uai", "ing"), true
		}
		return preferredFinals(initial, "ing", "uai"), true
	case 'l':
		if initial == "j" || initial == "q" || initial == "x" || initial == "n" || initial == "l" {
			return preferredFinals(initial, "iang", "uang"), true
		}
		return preferredFinals(initial, "uang", "iang"), true
	case 'x':
		if initial == "g" || initial == "k" || initial == "h" || initial == "zh" || initial == "ch" || initial == "sh" || initial == "r" || initial == "z" || initial == "c" || initial == "s" {
			return preferredFinals(initial, "ua", "ia"), true
		}
		return preferredFinals(initial, "ia", "ua"), true
	case 'v':
		if initial == "j" || initial == "q" || initial == "x" || initial == "y" {
			return preferredFinals(initial, "u", "v", "ui"), true
		}
		if initial == "n" || initial == "l" {
			return preferredFinals(initial, "v", "u", "ui"), true
		}
		return preferredFinals(initial, "ui", "u", "v"), true
	case 't':
		if initial == "n" || initial == "l" {
			return preferredFinals(initial, "ve", "ue"), true
		}
		if initial == "j" || initial == "q" || initial == "x" {
			return preferredFinals(initial, "ue", "ve"), true
		}
		return preferredFinals(initial, "ve", "ue"), true
	case 'r':
		if initial == "" {
			return []string{"er"}, true
		}
		return preferredFinals(initial, "uan"), true
	}
	fin, ok := xiaoheFinals[key]
	if !ok {
		return nil, false
	}
	return preferredFinals(initial, fin), true
}

func preferredFinals(initial string, fins ...string) []string {
	out := make([]string, 0, len(fins))
	seen := map[string]struct{}{}
	for _, fin := range fins {
		syl := fin
		if initial != "" {
			syl = initial + fin
		}
		if initial != "" && !IsValidSyllable(syl) {
			continue
		}
		if initial == "" && fin != "" && !IsValidSyllable(fin) {
			continue
		}
		if _, dup := seen[fin]; dup {
			continue
		}
		seen[fin] = struct{}{}
		out = append(out, fin)
	}
	if len(out) == 0 && len(fins) > 0 {
		return []string{fins[0]}
	}
	return out
}

func FlypyToFull(input string) string {
	syllables, rem, ok := ExpandXiaohe(input)
	if !ok && len(syllables) == 0 {
		return input
	}
	out := strings.Join(syllables, "")
	if rem != "" {
		out += rem
	}
	return out
}
