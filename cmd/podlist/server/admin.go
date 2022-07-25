package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/AppsFlyer/go-sundheit"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/AppsFlyer/go-sundheit/checks"
	healthhttp "github.com/AppsFlyer/go-sundheit/http"
)

func defaultAdminServer() *http.Server {
	cfg, _ := rest.InClusterConfig()
	clientset, _ := kubernetes.NewForConfig(cfg)
	h := gosundheit.New()

	h.RegisterCheck(
		&checks.CustomCheck{
			CheckName: "k8s-controlplane-healthz",
			CheckFunc: func(ctx context.Context) (details interface{}, err error) {
				result := clientset.Discovery().RESTClient().Get().AbsPath("/healthz").Do(ctx)
				b, _ := result.Raw()
				return string(b), result.Error()
			},
		},
		gosundheit.ExecutionPeriod(5*time.Second),
		gosundheit.ExecutionTimeout(time.Second),
	)

	mux := http.NewServeMux()
	mux.Handle("/healthz", healthhttp.HandleHealthJSON(h))
	// mux.Handle("/metrics", promhttp.Handler())

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

	return adminSrv
}
