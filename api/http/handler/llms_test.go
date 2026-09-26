package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLlmsTxt(t *testing.T) {
	h := &Handler{}

	req := httptest.NewRequest(http.MethodGet, "/llms.txt", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Header().Get("Content-Type"), "text/plain")
	require.Contains(t, rr.Body.String(), "GET /api/system/fork-capabilities")
	require.Contains(t, rr.Body.String(), "GET /api/docs/openapi.yaml")
	require.Contains(t, rr.Body.String(), "POST /api/stacks/create/standalone/string")
	require.Contains(t, rr.Body.String(), "POST /api/stacks/create/swarm/string")
	require.Contains(t, rr.Body.String(), "offline")

	req = httptest.NewRequest(http.MethodPost, "/llms.txt", nil)
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	require.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestOpenAPISpec(t *testing.T) {
	h := &Handler{}

	req := httptest.NewRequest(http.MethodGet, "/api/docs/openapi.yaml", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Header().Get("Content-Type"), "yaml")
	require.Contains(t, rr.Body.String(), `swagger: "2.0"`)
	require.Contains(t, rr.Body.String(), "basePath: /api")
	require.Contains(t, rr.Body.String(), "/stacks/create/standalone/string")
	require.Contains(t, rr.Body.String(), "/stacks/create/swarm/string")
	require.Contains(t, rr.Body.String(), "/stacks/{id}/git/redeploy")

	req = httptest.NewRequest(http.MethodPost, "/api/docs/openapi.yaml", nil)
	rr = httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	require.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}
