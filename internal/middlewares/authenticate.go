package middlewares

import (
	"context"
	"log"
	"net/http"
	"strings"

	"golangchatapp/internal/utils"
)

// const block is used to define a set of related constants. Prevents repetition
// and makes the code cleaner and easier to maintain. In this case, we are defining
// constants for context keys and platform types.
const (
	CtxUserID          string = "userId"
	CtxUserDisplayName string = "name"
	CtxPlatform        string = "X-Platform"
	CtxAuthorization   string = "Authorization"
	PlatformWeb               = "web"
	PlatformMobile            = "mobile"
)

/*
  - HandlerFunc is a type that represents a function that can handle HTTP
    requests. It takes an http.ResponseWriter and an *http.Request as parameters. This is the standard

*
*/
func Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := strings.TrimSpace(r.Header.Get(string(CtxAuthorization)))
		if authHeader == "" || !strings.HasPrefix(strings.ToLower(authHeader),
			"bearer ") {
			utils.JSON(w, http.StatusUnauthorized, false, "U0 - Unauthorized", nil)
			return
		}

		platform := strings.ToLower(strings.TrimSpace(r.Header.Get(string(CtxPlatform))))
		if platform != PlatformWeb && platform != PlatformMobile {
			utils.JSON(w, http.StatusUnauthorized, false, "invalid platform", nil)
			return
		}

		accessToken := strings.TrimSpace(authHeader[7:])
		userId, name, tokenPlatform, err := utils.VerifyJWT(accessToken)
		if err != nil {
			log.Println(err)
			utils.JSON(w, http.StatusUnauthorized, false, "U1 - Unauthorized", nil)
			return
		}

		if tokenPlatform != platform {
			utils.JSON(w, http.StatusUnauthorized, false, "U2 - Unauthorized", nil)
			return
		}

		// similar to NodeJs req.user, we are adding the userId and name to the request context so that it can be accessed in the next handler.
		ctx := r.Context()
		ctx = context.WithValue(ctx, CtxUserID, userId)
		ctx = context.WithValue(ctx, CtxUserDisplayName, name)
		ctx = context.WithValue(ctx, CtxPlatform, tokenPlatform)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
