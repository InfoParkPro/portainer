package audit

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/apikey"
	"github.com/portainer/portainer/api/datastore"
	"github.com/portainer/portainer/api/dataservices"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/jwt"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"

	"github.com/gorilla/mux"
	"github.com/segmentio/encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAuditor(t *testing.T) (*datastore.Store, *mux.Router, string) {
	t.Helper()

	_, store := datastore.MustNewTestStore(t, true, true)

	adminUser := &portainer.User{ID: 1, Username: "admin", Role: portainer.AdministratorRole}
	require.NoError(t, store.User().Create(adminUser))

	endpoint := &portainer.Endpoint{ID: 3, Name: "my-environment"}
	require.NoError(t, store.Endpoint().Create(endpoint))

	stack := &portainer.Stack{ID: 1, Name: "web", EndpointID: 3, AutoUpdate: &portainer.AutoUpdateSettings{Webhook: "8dce8c2f-9ca1-482b-ad20-271e86536ada"}}
	require.NoError(t, store.Stack().Create(stack))

	jwtService, err := jwt.NewService("1h", store)
	require.NoError(t, err)
	apiKeyService := apikey.NewAPIKeyService(store.APIKeyRepository(), store.User())
	requestBouncer := security.NewRequestBouncer(t.Context(), store, jwtService, apiKeyService)

	jwtToken, _, err := jwtService.GenerateToken(&portainer.TokenData{ID: adminUser.ID, Username: adminUser.Username, Role: adminUser.Role})
	require.NoError(t, err)

	router := mux.NewRouter()
	router.Handle("/test/{id}",
		requestBouncer.AuthenticatedAccess(
			Middleware(func() dataservices.DataStore { return store }, "test_action", "Field")(
				httperror.LoggerHandler(testNextHandler)),
		)).Methods(http.MethodPost)

	return store, router, jwtToken
}

var nextResult *httperror.HandlerError

func testNextHandler(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	if nextResult != nil {
		return nextResult
	}
	w.WriteHeader(http.StatusOK)
	return nil
}

func Test_Middleware(t *testing.T) {
	t.Run("records successful action with endpoint context from route id", func(t *testing.T) {
		store, router, jwtToken := setupAuditor(t)
		nextResult = nil

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/test/3", strings.NewReader(`{"Field":"value","Other":"ignored"}`))
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwtToken))
		router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)

		logs, err := store.UserActivityLog().Logs()
		require.NoError(t, err)
		require.Len(t, logs, 1)

		entry := logs[0]
		assert.Equal(t, "admin", entry.Username)
		assert.Equal(t, "my-environment", entry.Context)
		assert.Equal(t, "test_action", entry.Action)

		var payload map[string]any
		require.NoError(t, json.Unmarshal(entry.Payload, &payload))
		assert.Equal(t, true, payload["success"])
		assert.Equal(t, "value", payload["Field"])
		assert.NotContains(t, payload, "Other")
		assert.Equal(t, "3", payload["id"])
	})

	t.Run("resolves endpoint from stack webhook route variable", func(t *testing.T) {
		// covered indirectly by the stacks route wiring; here we check the
		// webhookID lookup helper directly
		store, _, _ := setupAuditor(t)

		stack, err := store.Stack().StackByWebhookID("8dce8c2f-9ca1-482b-ad20-271e86536ada")
		require.NoError(t, err)
		require.Equal(t, portainer.EndpointID(3), stack.EndpointID)
	})

	t.Run("records failed action with error message", func(t *testing.T) {
		store, router, jwtToken := setupAuditor(t)
		nextResult = &httperror.HandlerError{StatusCode: http.StatusBadRequest, Message: "something failed"}

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/test/999", strings.NewReader(`{}`))
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwtToken))
		router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)

		logs, err := store.UserActivityLog().Logs()
		require.NoError(t, err)
		require.Len(t, logs, 1)

		var payload map[string]any
		require.NoError(t, json.Unmarshal(logs[0].Payload, &payload))
		assert.Equal(t, false, payload["success"])
		assert.Equal(t, "something failed", payload["error"])
		assert.Equal(t, "environment 999", logs[0].Context)
	})
}
