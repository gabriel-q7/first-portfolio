package middleware

import "net/http"

func setSecureHeaders(w http.ResponseWriter) {
	h := w.Header()
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("X-Frame-Options", "DENY")
	h.Set("X-XSS-Protection", "1; mode=block")
	h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
	h.Set("Content-Security-Policy", "default-src 'self'")
	h.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
	h.Del("Server")
}

// SecureHeaders adds security-related HTTP response headers.
func SecureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setSecureHeaders(w)
		next.ServeHTTP(w, r)
	})
}

// SecureHeadersProd is like SecureHeaders but also adds HSTS (for production).
func SecureHeadersProd(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setSecureHeaders(w)
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		next.ServeHTTP(w, r)
	})
}
