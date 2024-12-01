package utils

import (
	"math/rand"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

var seededRand = rand.New(rand.NewSource(time.Now().UnixNano()))

func GenId() string {

	id := make([]byte, 6)
	for i := range id {
		id[i] = charset[seededRand.Intn(len(charset))]

	}
	return string(id)

}
