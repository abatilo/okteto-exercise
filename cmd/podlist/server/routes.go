package server

import (
	"net/http"
	"sort"
	"time"

	"github.com/abatilo/okteto-exercise/internal"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/hako/durafmt"
)

func (s *Server) RegisterRoutes(r *chi.Mux) {
	r.Get("/", s.index())
	r.Get("/api/v1/pods", s.listPods())
}

func (s *Server) index() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render.PlainText(w, r, "Hello, world!")
	}
}

func (s *Server) listPods() http.HandlerFunc {
	count := s.metrics.NewGauge(internal.GaugeOpts{
		Name: "podlist_pod_count",
		Help: "The total number of pods being listed",
	})

	type pod struct {
		Name     string `json:"name"`
		Restarts int32  `json:"restarts"`
		Age      string `json:"age"`
		ageInMS  int64  `json:"-"`
	}

	type response struct {
		Pods []pod `json:"pods"`
	}

	const (
		SortByName = iota
		SortByRestarts
		SortByAge
	)

	return func(w http.ResponseWriter, r *http.Request) {
		sortBy := SortByName
		sortParam := r.URL.Query().Get("sort")

		s.log.Debug().Str("sort", sortParam).Msg("Sort method")

		if sortParam == "restarts" {
			sortBy = SortByRestarts
		} else if sortParam == "age" {
			sortBy = SortByAge
		}

		podList, err := s.kubernetesClient.ListPods(r.Context())
		if err != nil {
			s.log.Error().Err(err).Msg("failed to list pods")
			render.Status(r, http.StatusInternalServerError)
			return
		}

		pods := make([]pod, len(podList.Items))
		for i, p := range podList.Items {
			totalRestarts := int32(0)
			for _, cs := range p.Status.ContainerStatuses {
				totalRestarts += cs.RestartCount
			}

			creationTime := p.GetCreationTimestamp().Time

			pods[i] = pod{
				Name:     p.Name,
				Restarts: totalRestarts,
				Age:      durafmt.Parse(time.Since(creationTime)).LimitFirstN(2).String(),
				ageInMS:  time.Since(creationTime).Milliseconds(),
			}
		}

		if sortBy == SortByName {
			sort.Slice(pods, func(i, j int) bool {
				return pods[i].Name < pods[j].Name
			})
		} else if sortBy == SortByRestarts {
			sort.Slice(pods, func(i, j int) bool {
				return pods[i].Restarts < pods[j].Restarts
			})
		} else if sortBy == SortByAge {
			sort.Slice(pods, func(i, j int) bool {
				return pods[i].ageInMS < pods[j].ageInMS
			})
		}

		resp := response{
			Pods: pods,
		}

		count.Set(float64(len(pods)))
		render.JSON(w, r, resp)
	}
}
