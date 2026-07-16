package utils

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	// make is a built-in function in Go that allocates and initializes a slice,
	// map, or channel. In this case, it creates a byte slice of length 32.
	// slice is a dynamically-sized, flexible view into the elements of an array.
	// In this case, it is used to hold 32 random bytes that will be generated
	// for the refresh token.

	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}
