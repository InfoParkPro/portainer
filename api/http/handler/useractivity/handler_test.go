package useractivity

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/apikey"
	"github.com/portainer/portainer/api/datastore"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/jwt"

	"github.com/segmentio/encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testContext struct {
	handler          *Handler
	store            *datastore.Store
	adminJWT         string
	standardJWT      string
}

func setupHandler(t *testing.T) *testContext {
	t.Helper()

	_, store := datastore.MustNewTestStore(t, true, true)

	adminUser := &portainer.User{ID: 1, Username: "admin", Role: portainer.AdministratorRole}
	require.NoError(t, store.User().Create(adminUser))

	standardUser := &portainer.User{ID: 2, Username: "standard", Role: portainer.StandardUserRole}
	require.NoError(t, store.User().Create(standardUser))

	jwtService, err := jwt.NewService("1h", store)
	require.NoError(t, err)
	apiKeyService := apikey.NewAPIKeyService(store.APIKeyRepository(), store.User())
	requestBouncer := security.NewRequestBouncer(t.Context(), store, jwtService, apiKeyService)

	handler := NewHandler(requestBouncer)
	handler.DataStore = store

	adminJWT, _, err := jwtService.GenerateToken(&portainer.TokenData{ID: adminUser.ID, Username: adminUser.Username, Role: adminUser.Role})
	require.NoError(t, err)
	standardJWT, _, err := jwtService.GenerateToken(&portainer.TokenData{ID: standardUser.ID, Username: standardUser.Username, Role: standardUser.Role})
	require.NoError(t, err)

	return &testContext{handler: handler, store: store, adminJWT: adminJWT, standardJWT: standardJWT}
}

func Test_activityLogs(t *testing.T) {
	ctx := setupHandler(t)

	entries := []*portainer.UserActivityLog{
		{ID: 1, Timestamp: 1710000000, Username: "admin", Context: "portainer", Action: "stack_create", Payload: []byte(`{"success":true}`)},
		{ID: 2, Timestamp: 1710000100, Username: "ci-bot", Context: "my-environment", Action: "image_pull", Payload: []byte(`{"success":true,"image":"nginx"}`)},
		{ID: 3, Timestamp: 1710000200, Username: "admin", Context: "portainer", Action: "stack_delete", Payload: []byte(`{"success":false,"error":"boom"}`)},
	}
	for _, entry := range entries {
		require.NoError(t, ctx.store.UserActivityLog().Log(entry))
	}

	t.Run("lists all entries newest first", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/useractivity/logs?sortDesc=true", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ctx.adminJWT))
		ctx.handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)

		var response userActivityLogsResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
		assert.Equal(t, 3, response.TotalCount)
		require.Len(t, response.Logs, 3)
		assert.Equal(t, int64(1710000200), response.Logs[0].Timestamp)
		assert.Equal(t, base64.StdEncoding.EncodeToString([]byte(`{"success":true}`)), response.Logs[2].Payload)
	})

	t.Run("filters by keyword", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/useractivity/logs?keyword=image_pull", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ctx.adminJWT))
		ctx.handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)

		var response userActivityLogsResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
		assert.Equal(t, 1, response.TotalCount)
		assert.Equal(t, "image_pull", response.Logs[0].Action)
	})

	t.Run("filters by time range in unix seconds", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/useractivity/logs?after=1710000101&before=1710000199", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ctx.adminJWT))
		ctx.handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)

		var response userActivityLogsResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
		assert.Equal(t, 0, response.TotalCount)
	})

	t.Run("paginates", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/useractivity/logs?offset=1&limit=1", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ctx.adminJWT))
		ctx.handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)

		var response userActivityLogsResponse
		require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &response))
		assert.Equal(t, 3, response.TotalCount)
		require.Len(t, response.Logs, 1)
		assert.Equal(t, int64(1710000100), response.Logs[0].Timestamp)
	})

	t.Run("denies non-admin", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/useractivity/logs", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ctx.standardJWT))
		ctx.handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusForbidden, rr.Code)
	})
}

func Test_activityLogsExport(t *testing.T) {
	ctx := setupHandler(t)

	entry := &portainer.UserActivityLog{ID: 1, Timestamp: 1710000000, Username: "admin", Context: "my-environment", Action: "stack_update", Payload: []byte(`{"success":true}`)}
	require.NoError(t, ctx.store.UserActivityLog().Log(entry))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/useractivity/logs.csv", nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ctx.adminJWT))
	ctx.handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Header().Get("Content-Disposition"), "attachment; filename=portainer-user-activity-logs-")
	assert.Contains(t, rr.Body.String(), "date,username,environment,action,api-key,success,payload")
	assert.Contains(t, rr.Body.String(), "admin,my-environment,stack_update,,true")
}
