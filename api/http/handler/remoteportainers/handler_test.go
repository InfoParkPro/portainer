package remoteportainers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/datastore"
	"github.com/portainer/portainer/api/internal/testhelpers"

	"github.com/stretchr/testify/require"
)

func setupHandler(t *testing.T) (*Handler, *datastore.Store) {
	t.Helper()

	_, store := datastore.MustNewTestStore(t, true, true)
	return NewHandler(testhelpers.NewTestRequestBouncer(), store), store
}

func TestCreateRemotePortainerHidesToken(t *testing.T) {
	t.Parallel()

	handler, store := setupHandler(t)

	body := bytes.NewBufferString(`{"Name":"standby","URL":"https://standby.example","APIToken":"ptr_token","TLSSkipVerify":true}`)
	request := httptest.NewRequest(http.MethodPost, "/remote_portainers", body)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)

	var response portainer.RemotePortainer
	require.NoError(t, json.NewDecoder(recorder.Body).Decode(&response))
	require.Equal(t, "standby", response.Name)
	require.Empty(t, response.APIToken)

	stored, err := store.RemotePortainer().Read(response.ID)
	require.NoError(t, err)
	require.Equal(t, "ptr_token", stored.APIToken)
}

func TestUpdateRemotePortainerKeepsExistingTokenWhenOmitted(t *testing.T) {
	t.Parallel()

	handler, store := setupHandler(t)
	remote := &portainer.RemotePortainer{
		Name:     "standby",
		URL:      "https://standby.example",
		APIToken: "ptr_token",
	}
	require.NoError(t, store.RemotePortainer().Create(remote))

	body := bytes.NewBufferString(`{"Name":"standby-renamed","URL":"https://standby2.example","TLSSkipVerify":true}`)
	request := httptest.NewRequest(http.MethodPut, "/remote_portainers/"+strconv.Itoa(int(remote.ID)), body)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)

	stored, err := store.RemotePortainer().Read(remote.ID)
	require.NoError(t, err)
	require.Equal(t, "standby-renamed", stored.Name)
	require.Equal(t, "https://standby2.example", stored.URL)
	require.Equal(t, "ptr_token", stored.APIToken)
	require.True(t, stored.TLSSkipVerify)
}

func TestRemoteStackUpdateUsesRemoteEndpointID(t *testing.T) {
	t.Parallel()

	var updateCalled bool
	remoteServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "ptr_token", r.Header.Get("X-API-KEY"))

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/stacks/7":
			require.NoError(t, json.NewEncoder(w).Encode(portainer.Stack{
				ID:         7,
				Name:       "report",
				Type:       portainer.DockerComposeStack,
				EndpointID: 42,
			}))
		case r.Method == http.MethodPut && r.URL.Path == "/api/stacks/7":
			updateCalled = true
			require.Equal(t, "42", r.URL.Query().Get("endpointId"))
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
	defer remoteServer.Close()

	handler, store := setupHandler(t)
	remote := &portainer.RemotePortainer{
		Name:     "standby",
		URL:      remoteServer.URL,
		APIToken: "ptr_token",
	}
	require.NoError(t, store.RemotePortainer().Create(remote))

	body := bytes.NewBufferString(`{"StackFileContent":"services: {}","Prune":true,"RepullImageAndRedeploy":true}`)
	request := httptest.NewRequest(http.MethodPut, "/remote_portainers/"+strconv.Itoa(int(remote.ID))+"/stacks/7", body)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.True(t, updateCalled)
}

func TestRemotePortainerTestConnectionReturnsBadGatewayOnRemoteError(t *testing.T) {
	t.Parallel()

	remoteServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad token", http.StatusUnauthorized)
	}))
	defer remoteServer.Close()

	handler, store := setupHandler(t)
	remote := &portainer.RemotePortainer{
		Name:     "standby",
		URL:      remoteServer.URL,
		APIToken: "ptr_token",
	}
	require.NoError(t, store.RemotePortainer().Create(remote))

	request := httptest.NewRequest(http.MethodPost, "/remote_portainers/"+strconv.Itoa(int(remote.ID))+"/test", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusBadGateway, recorder.Code)
	require.Contains(t, recorder.Body.String(), "bad token")
}
