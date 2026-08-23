package server

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net"
	"net/http"
	"net/url"
	"strings"
)

func randomToken() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}

func (server *Server) securityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Cache-Control", "no-store, max-age=0")
		writer.Header().Set("Pragma", "no-cache")
		writer.Header().Set("X-Content-Type-Options", "nosniff")
		writer.Header().Set("X-Frame-Options", "DENY")
		writer.Header().Set("Referrer-Policy", "no-referrer")
		writer.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		writer.Header().Set("Cross-Origin-Resource-Policy", "same-origin")
		writer.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; connect-src 'self'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'; form-action 'self'")

		if !loopbackHost(request.Host) {
			writeAPIError(writer, http.StatusBadRequest, "INVALID_HOST", "Request host must be loopback.")
			return
		}
		origin := request.Header.Get("Origin")
		if origin != "" && !server.allowedOrigin(origin) {
			writeAPIError(writer, http.StatusForbidden, "INVALID_ORIGIN", "Request origin is not allowed.")
			return
		}
		if origin != "" {
			writer.Header().Set("Access-Control-Allow-Origin", origin)
			writer.Header().Set("Vary", "Origin")
		}
		if request.Method == http.MethodOptions {
			writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
			writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-PlanParity-Token, Last-Event-ID")
			writer.WriteHeader(http.StatusNoContent)
			return
		}
		if requiresToken(request) && !constantTimeEqual(request.Header.Get("X-PlanParity-Token"), server.token) {
			writeAPIError(writer, http.StatusUnauthorized, "INVALID_SESSION", "A valid local session token is required.")
			return
		}
		next.ServeHTTP(writer, request)
	})
}

func loopbackHost(value string) bool {
	host, _, err := net.SplitHostPort(value)
	if err != nil {
		host = value
	}
	host = strings.Trim(host, "[]")
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func (server *Server) allowedOrigin(value string) bool {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme != "http" || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return false
	}
	allowed := map[string]bool{"http://" + server.Address(): true}
	for _, origin := range server.options.AllowedOrigins {
		allowed[origin] = true
	}
	return allowed[value]
}

func requiresToken(request *http.Request) bool {
	if !strings.HasPrefix(request.URL.Path, "/api/v1/") {
		return false
	}
	if request.Method == http.MethodGet {
		switch request.URL.Path {
		case "/api/v1/health", "/api/v1/system", "/api/v1/engines":
			return false
		}
	}
	return true
}

func constantTimeEqual(first, second string) bool {
	if len(first) != len(second) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(first), []byte(second)) == 1
}
