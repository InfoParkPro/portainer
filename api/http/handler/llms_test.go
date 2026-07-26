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
	require.Contains(t, rr.Body.String(), "offline")
}
