package gorm

import (
	"crypto/sha256"
	"fmt"
)

var HashSalt = "1Y3UUwS=Zg7eEtgkr9xHMZ5np9Tb0l6XOT6Xhx5km${q<hBCWiVZ=rA"

func getHexHash(v string, salt string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(v+salt)))
}
