package middleware

import (
	"crypto/subtle"
	"net/http"
	"os"
)

// BasicAuth implements HTTP Basic Authentication middleware
func BasicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := os.Getenv("SQLITE_REST_USERNAME")
		password := os.Getenv("SQLITE_REST_PASSWORD")

		// If authentication is not configured, skip authentication
		if username == "" || password == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Get credentials from request
		user, pass, ok := r.BasicAuth()
		if !ok {
			unauthorized(w)
			return
		}

		// Constant time comparison to prevent timing attacks
		usernameMatch := subtle.ConstantTimeCompare([]byte(user), []byte(username)) == 1
		passwordMatch := subtle.ConstantTimeCompare([]byte(pass), []byte(password)) == 1

		if !usernameMatch || !passwordMatch {
			unauthorized(w)
			return
		}

		// Authentication successful, call the next handler
		next.ServeHTTP(w, r)
	})
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte("Unauthorized"))
}
