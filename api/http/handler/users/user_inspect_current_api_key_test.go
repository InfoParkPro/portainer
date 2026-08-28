package users

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/datastore"
	"github.com/portainer/portainer/api/internal/testhelpers"

	"github.com/segmentio/encoding/json"
	"github.com/stretchr/testify/require"
)

func Test_userInspectCurrentAPIKey(t *testing.T) {
	t.Parallel()

	_, store := datastore.MustNewTestStore(t, true, true)

	user := &portainer.User{ID: 2, Username: "standard", Role: portainer.StandardUserRole}
	err := store.User().Create(user)
	require.NoError(t, err)

	h, jwtService, apiKeyService := newTestHandler(t, store)

	t.Run("returns the API key used by the current request", func(t *testing.T) {
		rawAPIKey, apiKey, err := apiKeyService.GenerateApiKey(*user, "current-key")
		require.NoError(t, err)

		apiKey.AccessPreset = portainer.APIKeyAccessPresetPower
		apiKey.TemporaryAccessPreset = portainer.APIKeyAccessPresetManage
		apiKey.TemporaryAccessExpiresAt = time.Now().UTC().Add(time.Hour).Unix()
		require.NoError(t, apiKeyService.UpdateAPIKey(apiKey))

		req := httptest.NewRequest(http.MethodGet, "/users/me/current-api-key", nil)
		req.Header.Add("x-api-key", rawAPIKey)

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)

		body, err := io.ReadAll(rr.Body)
		require.NoError(t, err)

		var resp apiKeyResponse
		err = json.Unmarshal(body, &resp)
		require.NoError(t, err)

		require.Equal(t, apiKey.ID, resp.ID)
		require.Equal(t, apiKey.UserID, resp.UserID)
		require.Equal(t, apiKey.Prefix, resp.Prefix)
		require.Equal(t, apiKey.Description, resp.Description)
		require.Empty(t, resp.Digest)
		require.Equal(t, portainer.APIKeyAccessPresetPower, resp.AccessPreset)
		require.Equal(t, portainer.APIKeyAccessPresetManage, resp.TemporaryAccessPreset)
		require.Equal(t, portainer.APIKeyAccessPresetManage, resp.EffectiveAccessPreset)
	})

	t.Run("rejects cookie or JWT authentication", func(t *testing.T) {
		jwt, _, err := jwtService.GenerateToken(&portainer.TokenData{ID: user.ID, Username: user.Username, Role: user.Role})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/users/me/current-api-key", nil)
		testhelpers.AddTestSecurityCookie(req, jwt)

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
