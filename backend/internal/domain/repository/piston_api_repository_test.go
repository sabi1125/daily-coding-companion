package repository

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/config"
	"backend/internal/domain/entities"
	logger "backend/internal/log"
	"backend/internal/response"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	logger.InitForTest()
	m.Run()
}

func TestPistonApiRepository_RunSubmission(t *testing.T) {
	tests := []struct {
		name           string
		buildServer    func() *httptest.Server
		wantStatusCode int
	}{
		{
			name: "success",
			buildServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "/v2/execute", r.URL.Path)
					assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
					w.Header().Set("Content-Type", "application/json")
					w.Header().Set("X-Auth", "47f88c9fe58a6861c6f7da8f5a44e5b30fa1b7f293355fe6391aacbc9193e3ec")
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"language":"python","version":"3.12.0","run":{"stdout":"1\n","stderr":"","code":0}}`))
				}))
			},
		},
		{
			name: "piston reached but returned an error body — bad gateway",
			buildServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`not json`))
				}))
			},
			wantStatusCode: http.StatusBadGateway,
		},
		{
			name: "piston rejects the request — non-2xx status maps to bad gateway, not a silent zero-valued 200",
			buildServer: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusBadRequest)
					w.Header().Set("X-Auth", "47f88c9fe58a6861c6f7da8f5a44e5b30fa1b7f293355fe6391aacbc9193e3ec")
					w.Write([]byte(`{"message":"language not found: python-2.0.0"}`))
				}))
			},
			wantStatusCode: http.StatusBadGateway,
		},
		{
			name: "piston unreachable — service unavailable",
			buildServer: func() *httptest.Server {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
				server.Close() // closed before use: connection refused, request never reaches Piston
				return server
			},
			wantStatusCode: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.buildServer()
			defer server.Close()

			repo := NewPistonApiRepository(&config.PistonConfig{PistonBaseApi: server.URL})

			res, err := repo.RunSubmission(context.Background(), entities.SubmittedSolutionRequest{
				Language: "python",
				Version:  "3.12.0",
				Files:    []entities.PistonFile{{Content: "print(1)"}},
			})

			if tt.wantStatusCode != 0 {
				var appErr *response.AppError
				ok := errors.As(err, &appErr)
				assert.True(t, ok, "expected a *response.AppError, got %T: %v", err, err)
				assert.Equal(t, tt.wantStatusCode, appErr.Status.Code)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, "python", res.Language)
			assert.Equal(t, "1\n", res.Run.Stdout)
		})
	}
}
