package middlewares

import (
	"encoding/json"
	"log"
	"net/http"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[%s] %s %s from [%s]", r.Method, r.URL.Path, r.Proto, r.RemoteAddr)
		log.Printf("Headers: %v", r.Header)
		log.Printf("Query Params: %v", r.URL.Query())
		log.Printf("Request Body: %v", json.NewDecoder(r.Body).Decode(&struct{}{})) // just to read the body without storing it
		log.Printf("Request Cookies: %v", r.Cookies())
		log.Printf("Request Host: %v", r.Host)
		next.ServeHTTP(w, r)
	})
}
