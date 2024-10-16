package secure

import (
	"crypto/rand"
	"math/big"
)

func GenerateRandomBytes(n uint32) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}

	return b
}

const alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

func GenerateRandomString(n int) string {
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		ret[i] = alphabet[num.Int64()]
	}
	return string(ret)
}
