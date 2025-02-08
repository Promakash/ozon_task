package random

import (
	"crypto/rand"
	"math/big"
)

func NewRandomString(size int, alphabet string) (string, error) {
	b := make([]byte, size)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		b[i] = alphabet[num.Int64()]
	}

	return string(b), nil
}
