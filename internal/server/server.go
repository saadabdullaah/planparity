// Package server provides PlanParity's authenticated localhost-only API and UI.
package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/saadabdullaah/planparity/internal/app"
)

// Options controls the local listener and development origins.
type Options struct {
	Host              string
	Port              int
	AllowedOrigins    []string
	TerraformPath     string
	OpenTofuPath      string
	MaxRequestBytes   int64
	ReadHeaderTimeout time.Duration
}

// Server owns one process-local API, session token, job manager, and listener.
type Server struct {
	options  Options
	service  *app.Service
	jobs     *app.JobManager
	token    string
	handler  http.Handler
	http     *http.Server
	listener net.Listener
	mu       sync.RWMutex
	address  string
}

// New constructs a server without binding a port.
func New(service *app.Service, options Options) (*Server, error) {
	if service == nil {
		service = app.NewService(nil, 5*time.Minute)
	}
	if options.Host == "" {
		options.Host = "127.0.0.1"
	}
	ip := net.ParseIP(options.Host)
	if ip == nil || !ip.IsLoopback() {
		return nil, errors.New("PlanParity UI host must be a numeric loopback address")
	}
	if options.Port < 0 || options.Port > 65535 {
		return nil, errors.New("port must be between 0 and 65535")
	}
	if options.MaxRequestBytes <= 0 {
		options.MaxRequestBytes = 1 << 20
	}
	if options.ReadHeaderTimeout <= 0 {
		options.ReadHeaderTimeout = 5 * time.Second
	}
	token, err := randomToken()
	if err != nil {
		return nil, err
	}
	server := &Server{options: options, service: service, token: token}
	server.jobs = app.NewJobManager(context.Background(), 100)
	server.handler = server.securityMiddleware(server.routes())
	server.http = &http.Server{Handler: server.handler, ReadHeaderTimeout: options.ReadHeaderTimeout, IdleTimeout: 60 * time.Second}
	return server, nil
}

// Listen reserves the configured loopback port.
func (server *Server) Listen() error {
	server.mu.Lock()
	defer server.mu.Unlock()
	if server.listener != nil {
		return errors.New("server is already listening")
	}
	listener, err := net.Listen("tcp", net.JoinHostPort(server.options.Host, strconv.Itoa(server.options.Port)))
	if err != nil {
		return err
	}
	server.listener = listener
	server.address = listener.Addr().String()
	return nil
}

// Serve blocks until shutdown or a listener error.
func (server *Server) Serve(ctx context.Context) error {
	server.mu.RLock()
	listener := server.listener
	server.mu.RUnlock()
	if listener == nil {
		if err := server.Listen(); err != nil {
			return err
		}
		server.mu.RLock()
		listener = server.listener
		server.mu.RUnlock()
	}
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			shutdownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = server.http.Shutdown(shutdownContext)
		case <-done:
		}
	}()
	err := server.http.Serve(listener)
	close(done)
	server.jobs.Close()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

// Handler exposes the secured handler for tests.
func (server *Server) Handler() http.Handler { return server.handler }

// SessionToken returns the ephemeral token used by the CLI bootstrap URL.
func (server *Server) SessionToken() string { return server.token }

// Address returns the bound host and port after Listen.
func (server *Server) Address() string {
	server.mu.RLock()
	defer server.mu.RUnlock()
	if server.address != "" {
		return server.address
	}
	return net.JoinHostPort(server.options.Host, strconv.Itoa(server.options.Port))
}

// URL returns the browser bootstrap URL. The token is in the fragment and is not sent in HTTP requests.
func (server *Server) URL() string {
	return fmt.Sprintf("http://%s/#token=%s", server.Address(), server.token)
}
