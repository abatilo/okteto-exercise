package server

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Server struct {
	log         zerolog.Logger
	server      *http.Server
	adminServer *http.Server

	kubernetesClient *kubernetes.Clientset
}

// ServerOption lets you functionally control construction of the web server
type ServerOption func(s *Server)

func NewServer(log zerolog.Logger, options ...ServerOption) *Server {
	r := chi.NewRouter()
	s := &Server{
		log: log,
		server: &http.Server{
			Addr:    ":8080",
			Handler: r,
		},
	}

	// Overrides
	for _, option := range options {
		option(s)
	}

	// Defaults
	if s.kubernetesClient == nil {
		cfg, _ := rest.InClusterConfig()
		clientset, _ := kubernetes.NewForConfig(cfg)
		s.kubernetesClient = clientset
	}

	if s.adminServer == nil {
		s.adminServer = defaultAdminServer()
	}

	s.registerRoutes(r)
	return s
}

func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
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

func WithKubernetesClient(clientset *kubernetes.Clientset) ServerOption {
	return func(s *Server) {
		s.kubernetesClient = clientset
	}
}
