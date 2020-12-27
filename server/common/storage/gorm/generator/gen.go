package generator

import (
	"math/rand"
)

func GeneratePassword(lc, uc, dc, sc int) string {
	var (
		lowercaseLetters = getSymbolsSlice('a', 'z')
		uppercaseLetters = getSymbolsSlice('A', 'Z')
		digits           = getSymbolsSlice('0', '9')
		symbols          = []byte(`!$%()*+-<=>?@[]_{}`)
	)

	randBytes := make([]byte, 0)
	randBytes = append(randBytes, getRandomString(lowercaseLetters, lc)...)
	randBytes = append(randBytes, getRandomString(uppercaseLetters, uc)...)
	randBytes = append(randBytes, getRandomString(digits, dc)...)
	randBytes = append(randBytes, getRandomString(symbols, sc)...)
	rand.Shuffle(len(randBytes), func(i, j int) {
		randBytes[i], randBytes[j] = randBytes[j], randBytes[i]
	})
	return string(randBytes)
}

func getSymbolsSlice(start, end byte) []byte {
	out := make([]byte, end-start+1)
	for i := byte(0); start+i <= end; i++ {
		out[i] = start + i
	}
	return out
}

func getRandomString(symbols []byte, count int) []byte {
	out := make([]byte, count)
	for i := 0; i < count; i++ {
		si := rand.Intn(len(symbols))
		out[i] = symbols[si]
	}
	return out
}
