package session

const (
	defaultContextWindow = 48
	contextMax           = 64
)

func clampContextWindow(window int) int {
	if window <= 0 {
		return defaultContextWindow
	}
	if window > contextMax {
		return contextMax
	}
	return window
}

func appendCommittedRunes(dst []rune, text string, max int) []rune {
	if text == "" {
		return dst
	}
	if max <= 0 {
		max = contextMax
	}
	if max > contextMax {
		max = contextMax
	}
	for _, r := range text {
		dst = append(dst, r)
	}
	if len(dst) > max {
		dst = append([]rune(nil), dst[len(dst)-max:]...)
	}
	return dst
}

func unicodeTail(text string, window int) string {
	window = clampContextWindow(window)
	runess := []rune(text)
	if len(runess) <= window {
		return text
	}
	return string(runess[len(runess)-window:])
}

func joinCommittedAndConfirmed(committed []rune, confirmed string) string {
	if len(committed) == 0 {
		return confirmed
	}
	if confirmed == "" {
		return string(committed)
	}
	return string(committed) + confirmed
}
