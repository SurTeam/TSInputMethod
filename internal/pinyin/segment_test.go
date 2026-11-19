package pinyin

import "testing"

func TestSegmentContinuousNiHao(t *testing.T) {
	seg := NewSegmenter(nil)
	got := seg.PreferLongest("nihao")
	if len(got.Syllables) != 2 || got.Syllables[0] != "ni" || got.Syllables[1] != "hao" || got.Remainder != "" {
		t.Fatalf("got %#v", got)
	}
}

func TestSegmentPartialRemainder(t *testing.T) {
	seg := NewSegmenter(nil)
	got := seg.PreferLongest("nih")
	if len(got.Syllables) != 1 || got.Syllables[0] != "ni" || got.Remainder != "h" {
		t.Fatalf("got %#v", got)
	}
}

func TestSegmentXianAmbiguity(t *testing.T) {
	seg := NewSegmenter(nil)
	all := seg.Segment("xian")
	hasXian := false
	hasXiAn := false
	for _, item := range all {
		if len(item.Syllables) == 1 && item.Syllables[0] == "xian" && item.Remainder == "" {
			hasXian = true
		}
		if len(item.Syllables) == 2 && item.Syllables[0] == "xi" && item.Syllables[1] == "an" && item.Remainder == "" {
			hasXiAn = true
		}
	}
	if !hasXian || !hasXiAn {
		t.Fatalf("expected both xian and xi+an, got %#v", all)
	}
}
