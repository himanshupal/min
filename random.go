package main

import (
	"math/rand"
	"time"
)

const charset string = "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ" + "0123456789"

func RandomString(length int) string {
	random := rand.NewSource(time.Now().UnixNano())
	b := make([]byte, length)

	for i := range b {
		b[i] = charset[random.Int63()%int64(len(charset))]
	}

	return string(b)
}
