package system

import (
	"net/http"
	"net/http/httptest"
	"testing"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/internal/testhelpers"
	"github.com/segmentio/encoding/json"
	"github.com/stretchr/testify/require"
)

func TestForkCapabilities(t *testing.T) {
	h := NewHandler(testhelpers.NewTestRequestBouncer(), &portainer.Status{}, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/system/fork-capabilities", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Header().Get("Content-Type"), "application/json")

	var response struct {
		Fork          string `json:"fork"`
		Version       string `json:"version"`
		AccessPresets map[string]struct {
			Allowed []string `json:"allowed"`
			Denied  []string `json:"denied,omitempty"`
		} `json:"accessPresets"`
		Methods []struct {
			Method string `json:"method"`
			Path   string `json:"path"`
		} `json:"methods"`
	}

	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
	require.Equal(t, "infopark-portainer", response.Fork)
	require.Equal(t, portainer.APIVersion, response.Version)
	require.Contains(t, response.AccessPresets["power"].Allowed, "PUT /api/endpoints/{id}/forceupdateservice")
	require.Contains(t, response.AccessPresets["power"].Allowed, "POST /api/endpoints/{id}/docker/{version}/containers/{containerID}/exec when the container has label portainer.infopark.power.exec=true and passes safety checks")
	require.Contains(t, response.AccessPresets["power"].Allowed, "POST /api/endpoints/{id}/docker/{version}/exec/{execID}/start when the exec target container passes Power exec safety checks")
	require.Contains(t, response.AccessPresets["power"].Allowed, "GET /api/websocket/exec for exec sessions created through an allowed container exec request")
	require.Contains(t, response.AccessPresets["power"].Denied, "POST /api/endpoints/{id}/docker/{version}/services/{serviceID}/update")
	require.Contains(t, response.AccessPresets["power"].Denied, "Docker exec for containers without label portainer.infopark.power.exec=true")

	var foundForceUpdate bool
	var foundExecCreate bool
	for _, method := range response.Methods {
		if method.Method == http.MethodPut && method.Path == "/api/endpoints/{id}/forceupdateservice" {
			foundForceUpdate = true
		}
		if method.Method == http.MethodPost && method.Path == "/api/endpoints/{id}/docker/{version}/containers/{containerID}/exec" {
			foundExecCreate = true
		}
	}
	require.True(t, foundForceUpdate)
	require.True(t, foundExecCreate)
}
