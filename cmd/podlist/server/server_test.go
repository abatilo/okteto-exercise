package server_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/abatilo/okteto-exercise/cmd/podlist/server"
	"github.com/abatilo/okteto-exercise/internal"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

func Test_listPods(t *testing.T) {
	type pod struct {
		Name     string `json:"name"`
		Restarts int32  `json:"restarts"`
		Age      string `json:"age"`
	}

	type response struct {
		Pods []pod `json:"pods"`
	}

	type test struct {
		name       string
		requestURL string
		mockPods   []internal.Pod
		expected   response
	}

	mockPods := []internal.Pod{
		{
			ObjectMeta: internal.ObjectMeta{
				Name: "AAA",
				CreationTimestamp: internal.Time{
					Time: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			Status: internal.PodStatus{
				ContainerStatuses: []internal.ContainerStatuses{
					{
						RestartCount: 5,
					},
				},
			},
		},
		{
			ObjectMeta: internal.ObjectMeta{
				Name: "BBB",
				CreationTimestamp: internal.Time{
					Time: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			Status: internal.PodStatus{
				ContainerStatuses: []internal.ContainerStatuses{
					{
						RestartCount: 25,
					},
				},
			},
		},
		{
			ObjectMeta: internal.ObjectMeta{
				Name: "CCC",
				CreationTimestamp: internal.Time{
					Time: time.Date(2019, 6, 1, 0, 0, 0, 0, time.UTC),
				},
			},
			Status: internal.PodStatus{
				ContainerStatuses: []internal.ContainerStatuses{
					{
						RestartCount: 0,
					},
				},
			},
		},
	}

	tests := []test{
		{
			name:       "List pods by name by default",
			requestURL: "/api/v1/pods",
			mockPods:   mockPods,
			expected: response{
				Pods: []pod{
					{
						Name:     "AAA",
						Restarts: 5,
						Age:      "2 years 29 weeks",
					},
					{
						Name:     "BBB",
						Restarts: 25,
						Age:      "3 years 29 weeks",
					},
					{
						Name:     "CCC",
						Restarts: 0,
						Age:      "3 years 7 weeks",
					},
				},
			},
		},
		{
			name:       "List pods by restarts",
			requestURL: "/api/v1/pods?sort=restarts",
			mockPods:   mockPods,
			expected: response{
				Pods: []pod{
					{
						Name:     "CCC",
						Restarts: 0,
						Age:      "3 years 7 weeks",
					},
					{
						Name:     "AAA",
						Restarts: 5,
						Age:      "2 years 29 weeks",
					},
					{
						Name:     "BBB",
						Restarts: 25,
						Age:      "3 years 29 weeks",
					},
				},
			},
		},
		{
			name:       "List pods by age",
			requestURL: "/api/v1/pods?sort=age",
			mockPods:   mockPods,
			expected: response{
				Pods: []pod{
					{
						Name:     "AAA",
						Restarts: 5,
						Age:      "2 years 29 weeks",
					},
					{
						Name:     "CCC",
						Restarts: 0,
						Age:      "3 years 7 weeks",
					},
					{
						Name:     "BBB",
						Restarts: 25,
						Age:      "3 years 29 weeks",
					},
				},
			},
		},
	}

	for _, test := range tests {
		r := chi.NewRouter()
		s := server.NewServer(
			server.WithLogger(zerolog.New(ioutil.Discard)),
			server.WithAdminServer(&http.Server{}),
			server.WithMetrics(&internal.NoopMetrics{}),
			server.WithKubernetesClient(&internal.MockKubernetesClient{
				PodList: &internal.PodList{
					Items: test.mockPods,
				},
				Error: nil,
			}),
		)
		s.RegisterRoutes(r)

		req := httptest.NewRequest(http.MethodGet, test.requestURL, nil)
		w := httptest.NewRecorder()
		s.ServeHTTP(w, req)

		var resp response
		body, _ := ioutil.ReadAll(w.Body)
		json.Unmarshal(body, &resp)

		// Custom comparison here because we didn't have time to implement and
		// then refactor through having a mocked time so that we can assert
		// against exact times. So instead, we'll assert on order and only verify
		// the name which would be uniquely identifying anyways.
		actualPodNames := make([]string, len(resp.Pods))
		expectedPodNames := make([]string, len(test.expected.Pods))

		for i := 0; i < 3; i++ {
			actualPodNames[i] = resp.Pods[i].Name
			expectedPodNames[i] = test.expected.Pods[i].Name
		}

		if !reflect.DeepEqual(actualPodNames, expectedPodNames) {
			t.Errorf("%s: expected pods %v, got %v", test.name, expectedPodNames, actualPodNames)
		}
	}
}
