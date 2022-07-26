package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/AppsFlyer/go-sundheit"
	"github.com/AppsFlyer/go-sundheit/checks"
	healthhttp "github.com/AppsFlyer/go-sundheit/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (s *Server) defaultAdminServer() *http.Server {
	h := gosundheit.New()

	h.RegisterCheck(
		&checks.CustomCheck{
			CheckName: "k8s-controlplane-healthz",
			CheckFunc: func(ctx context.Context) (details interface{}, err error) {
				result := s.kubernetesClient.Healthz(ctx)
				b, _ := result.Raw()
				return string(b), result.Error()
			},
		},
		gosundheit.ExecutionPeriod(5*time.Second),
		gosundheit.ExecutionTimeout(time.Second),
	)

	mux := http.NewServeMux()
	mux.Handle("/healthz", healthhttp.HandleHealthJSON(h))
	mux.Handle("/metrics", promhttp.Handler())

	// pprof
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	adminSrv := &http.Server{
		Addr:    fmt.Sprintf(":8081"),
		Handler: mux,
	}

	go adminSrv.ListenAndServe()
	return adminSrv
}
