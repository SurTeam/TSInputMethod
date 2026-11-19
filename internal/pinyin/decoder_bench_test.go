package pinyin

import "testing"

func BenchmarkDecodeNihao(b *testing.B) {
	dec := NewDecoder(fixtureDict(), 16)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = dec.Decode("nihao")
	}
}

func BenchmarkDecodeShijie(b *testing.B) {
	dec := NewDecoder(fixtureDict(), 16)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = dec.Decode("shijie")
	}
}
