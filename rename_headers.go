package traefik_custom_headers_plugin

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
)

// Config holds the plugin configuration.
type Config struct{}

// CreateConfig creates and initializes the plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

// renameHeaders is the main plugin struct.
type renameHeaders struct {
	name string
	next http.Handler
}

// New creates a new Custom Header plugin.
func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	return &renameHeaders{
		name: name,
		next: next,
	}, nil
}

func (r *renameHeaders) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Get the value of CF-Connecting-IP
	cfConnectingIP := req.Header.Get("CF-Connecting-IP")

	// If CF-Connecting-IP exists, set it as the value of X-Forwarded-For
	if cfConnectingIP != "" {
		req.Header.Set("X-Forwarded-For", cfConnectingIP)
	}

	// Pass the request to the next handler
	r.next.ServeHTTP(rw, req)
}

type responseWriter struct {
	writer http.ResponseWriter
}

func (r *responseWriter) Header() http.Header {
	return r.writer.Header()
}

func (r *responseWriter) Write(bytes []byte) (int, error) {
	return r.writer.Write(bytes)
}

func (r *responseWriter) WriteHeader(statusCode int) {
	r.writer.WriteHeader(statusCode)
}

func (r *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := r.writer.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("%T is not a http.Hijacker", r.writer)
	}

	return hijacker.Hijack()
}

func (r *responseWriter) Flush() {
	if flusher, ok := r.writer.(http.Flusher); ok {
		flusher.Flush()
	}
}
