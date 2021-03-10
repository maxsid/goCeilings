package naming

import "fmt"

type NameIterator struct {
	start, end int32
	Counter    uint
}

func NewNameIterator(firstSymbol, lastSymbol int32) *NameIterator {
	return &NameIterator{start: firstSymbol, end: lastSymbol}
}

func (ni *NameIterator) Next() string {
	symbolsNum := uint(ni.end-ni.start) + 1
	s := string(ni.start + int32(ni.Counter%symbolsNum))
	if n := ni.Counter / symbolsNum; n != 0 {
		s = fmt.Sprintf("%s%d", s, n)
	}
	ni.Counter++
	return s
}
