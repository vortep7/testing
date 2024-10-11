package random

import (
	"math/rand"
)

const (
	availableChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func NewRandomString(size int) string {
	charSet := []byte(availableChars)
	alias := make([]byte, size)
	for i := range alias {
		alias[i] = charSet[rand.Intn(len(charSet))]
	}
	return string(alias)
}
