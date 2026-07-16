package utils

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
	hashBytes, err := bcrypt.GenerateFromPassword(
		[]byte(password), bcrypt.DefaultCost) // we use []byte and not ordinary
	// string because bcrypt works with byte slices. The DefaultCost is a
	// constant that defines the computational cost of hashing the password.
	// A higher cost means more security but also more time to compute the
	// hash.
	return string(hashBytes), err
}

func CheckHashPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
