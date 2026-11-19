package ranking

type Candidate struct {
	Text    string
	Comment string
	Score   int32
}

type Request struct {
	Pinyin     string
	Candidates []Candidate
	Context    string
	SessionID  string
	Generation uint64
}
