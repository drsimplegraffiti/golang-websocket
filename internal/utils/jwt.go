package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey []byte // byte: think of it as a slice of bytes, which is a more
// efficient way to handle binary data in Go.

func InitJwt(key string) {
	jwtKey = []byte(key)
}

type CustomClaims struct {
	UserID int64 `json:"user_id"` // the `json` tags specify how the fields
	// should be serialized to JSON. For example, the `UserID` field will be
	// serialized as "user_id" in the JSON output.
	Name     string `json:"name"`
	Platform string `json:"X-platform"`
	jwt.RegisteredClaims
}

func GenerateJWT(userId int64, name, platform string) (string, error) {
	exp := time.Now().Add(24 * time.Hour) // the expiration time is set to 24
	// hours from the current time
	if platform != "web" && platform != "mobile" {
		return "", errors.New("invalid platform for token")
	}

	claims := &CustomClaims{
		UserID:   userId,
		Name:     name,
		Platform: platform,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
			Subject:   fmt.Sprint(userId), // fmt.Sprint converts the userId to
			// a string and then holds it in the Subject field of the
			// RegisteredClaims struct. This is useful for identifying the user
			// associated with the token.
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func VerifyJWT(tokenStr string) (int64, string, string, error) {
	if len(jwtKey) == 0 {
		return 0, "", "", errors.New("jwt key not initialized")
	}

	token, err := jwt.ParseWithClaims(
		tokenStr,
		&CustomClaims{},
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return jwtKey, nil
		},
	)
	if err != nil {
		return 0, "", "", err
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return 0, "", "", errors.New("invalid claims type")
	}

	if claims.UserID == 0 || claims.Name == "" {
		return 0, "", "", errors.New("invalid user claims")
	}

	if claims.Platform != "web" && claims.Platform != "mobile" {
		return 0, "", "", errors.New("invalid platform claim")
	}

	return claims.UserID, claims.Name, claims.Platform, nil
}
