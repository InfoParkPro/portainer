package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	portainer "github.com/portainer/portainer/api"

	"github.com/stretchr/testify/require"
)

func TestClientSendsAPIKeyAndReadsStatus(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/status", r.URL.Path)
		require.Equal(t, "ptr_token", r.Header.Get("X-API-KEY"))

		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(StatusResponse{Version: "2.42.0"}))
	}))
	defer server.Close()

	client, err := New(server.URL, "ptr_token", false)
	require.NoError(t, err)

	status, err := client.Status(context.Background())
	require.NoError(t, err)
	require.Equal(t, "2.42.0", status.Version)
}

func TestClientUpdateStackUsesRemoteEndpointID(t *testing.T) {
	t.Parallel()

	var updateCalled bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "ptr_token", r.Header.Get("X-API-KEY"))

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/stacks/7":
			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(portainer.Stack{
				ID:         7,
				Name:       "report",
				Type:       portainer.DockerComposeStack,
				EndpointID: 42,
			}))
		case r.Method == http.MethodPut && r.URL.Path == "/api/stacks/7":
			updateCalled = true
			require.Equal(t, "42", r.URL.Query().Get("endpointId"))

			var payload UpdateStackPayload
			require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
			require.Equal(t, "services: {}", payload.StackFileContent)
			require.True(t, payload.Prune)
			require.True(t, payload.RepullImageAndRedeploy)

			w.Header().Set("Content-Type", "application/json")
			require.NoError(t, json.NewEncoder(w).Encode(portainer.Stack{
				ID:         7,
				Name:       "report",
				Type:       portainer.DockerComposeStack,
				EndpointID: 42,
			}))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := New(server.URL, "ptr_token", false)
	require.NoError(t, err)

	stack, err := client.Stack(context.Background(), 7)
	require.NoError(t, err)

	updated, err := client.UpdateStack(context.Background(), stack, UpdateStackPayload{
		StackFileContent:       "services: {}",
		Prune:                  true,
		RepullImageAndRedeploy: true,
	})
	require.NoError(t, err)
	require.True(t, updateCalled)
	require.Equal(t, portainer.StackID(7), updated.ID)
}

func TestClientReturnsRemoteErrorBody(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad token", http.StatusUnauthorized)
	}))
	defer server.Close()

	client, err := New(server.URL, "ptr_token", false)
	require.NoError(t, err)

	_, err = client.Status(context.Background())
	require.Error(t, err)
	require.Contains(t, err.Error(), "401")
	require.Contains(t, err.Error(), "bad token")
}
