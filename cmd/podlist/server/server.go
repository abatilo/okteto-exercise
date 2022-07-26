package server

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/abatilo/okteto-exercise/internal"
)

type Server struct {
	log         zerolog.Logger
	server      *http.Server
	adminServer *http.Server

	metrics          internal.MetricsClient
	kubernetesClient internal.ControlPlaneClient
}

// ServerOption lets you functionally control construction of the web server
type ServerOption func(s *Server)

func NewServer(options ...ServerOption) *Server {
	r := chi.NewRouter()
	s := &Server{
		server: &http.Server{
			Addr:    ":8080",
			Handler: r,
		},
	}

	// Overrides
	for _, option := range options {
		option(s)
	}

	s.RegisterRoutes(r)
	return s
}

func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.server.Handler.ServeHTTP(w, r)
}

func WithLogger(log zerolog.Logger) ServerOption {
	return func(s *Server) {
		s.log = log
	}
}

func WithAdminServer(adminServer *http.Server) ServerOption {
	return func(s *Server) {
		s.adminServer = adminServer
	}
}

func WithMetrics(metrics internal.MetricsClient) ServerOption {
	return func(s *Server) {
		s.metrics = metrics
	}
}

func WithKubernetesClient(clientset internal.ControlPlaneClient) ServerOption {
	return func(s *Server) {
		s.kubernetesClient = clientset
	}
}
