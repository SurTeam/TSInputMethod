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
}

func ExpandXiaohe(input string) (syllables []string, remainder string, ok bool) {
	input = strings.ToLower(strings.TrimSpace(input))
	if input == "" {
		return nil, "", true
	}
	for _, r := range input {
		if r < 'a' || r > 'z' {
			return nil, input, false
		}
	}
	i := 0
	for i+1 < len(input) {
		pair := input[i : i+2]
		syl, mapped := mapXiaohePair(pair[0], pair[1])
		if !mapped {
			return syllables, input[i:], false
		}
		syllables = append(syllables, syl)
		i += 2
	}
	if i < len(input) {
		return syllables, input[i:], true
	}
	return syllables, "", true
}

func mapXiaohePair(first, second byte) (string, bool) {
	if syl, ok := xiaoheZeroSyllables[string([]byte{first, second})]; ok {
		return syl, true
	}
	init, okInit := xiaoheInitials[first]
	if !okInit {
		return "", false
	}
	fin, okFin := resolveXiaoheFinal(init, second)
	if !okFin {
		return "", false
	}
	if init == "" {
		return fin, true
	}
	return init + fin, true
}

func resolveXiaoheFinal(initial string, key byte) (string, bool) {
	switch key {
	case 'o':
		if initial == "" || initial == "b" || initial == "p" || initial == "m" || initial == "f" {
			return "o", true
		}
		return "uo", true
	case 's':
		if initial == "j" || initial == "q" || initial == "x" {
			return "iong", true
		}
		return "ong", true
	case 'k':
		if initial == "g" || initial == "k" || initial == "h" || initial == "zh" || initial == "ch" || initial == "sh" || initial == "r" || initial == "z" || initial == "c" || initial == "s" {
			return "uai", true
		}
		return "ing", true
	case 'l':
		if initial == "j" || initial == "q" || initial == "x" || initial == "n" || initial == "l" {
			return "iang", true
		}
		return "uang", true
	case 'x':
		if initial == "g" || initial == "k" || initial == "h" || initial == "zh" || initial == "ch" || initial == "sh" || initial == "r" || initial == "z" || initial == "c" || initial == "s" {
			return "ua", true
		}
		return "ia", true
	case 'v':
		if initial == "j" || initial == "q" || initial == "x" || initial == "y" {
			return "u", true
		}
		if initial == "n" || initial == "l" {
			return "v", true
		}
		return "ui", true
	case 't':
		if initial == "j" || initial == "q" || initial == "x" || initial == "n" || initial == "l" {
			return "ue", true
		}
		return "ve", true
	case 'r':
		if initial == "" {
			return "er", true
		}
		return "uan", true
	}
	fin, ok := xiaoheFinals[key]
	return fin, ok
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
