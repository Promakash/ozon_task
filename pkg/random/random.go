package random

import (
	"crypto/rand"
	"math/big"
)

const allowedSymbols = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_"

func NewRandomString(size int) (string, error) {
	b := make([]byte, size)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(allowedSymbols))))
		if err != nil {
			return "", err
		}
		b[i] = allowedSymbols[num.Int64()]
	}

	return string(b), nil
}
