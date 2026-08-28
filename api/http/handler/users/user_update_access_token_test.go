package users

import (
	"bytes"
	"fmt"
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

func Test_userUpdateAccessToken(t *testing.T) {
	t.Parallel()

	_, store := datastore.MustNewTestStore(t, true, true)

	user := &portainer.User{ID: 2, Username: "standard", Role: portainer.StandardUserRole}
	err := store.User().Create(user)
	require.NoError(t, err)

	otherUser := &portainer.User{ID: 3, Username: "other", Role: portainer.StandardUserRole}
	err = store.User().Create(otherUser)
	require.NoError(t, err)

	h, jwtService, apiKeyService := newTestHandler(t, store)

	jwt, _, err := jwtService.GenerateToken(&portainer.TokenData{ID: user.ID, Username: user.Username, Role: user.Role})
	require.NoError(t, err)

	t.Run("user can update access token preset", func(t *testing.T) {
		_, apiKey, err := apiKeyService.GenerateApiKey(*user, "test-update-token")
		require.NoError(t, err)

		payload, err := json.Marshal(userAccessTokenUpdatePayload{AccessPreset: portainer.APIKeyAccessPresetReadOnly})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/users/%d/tokens/%d", user.ID, apiKey.ID), bytes.NewBuffer(payload))
		testhelpers.AddTestSecurityCookie(req, jwt)

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)

		body, err := io.ReadAll(rr.Body)
		require.NoError(t, err)

		var resp apiKeyResponse
		err = json.Unmarshal(body, &resp)
		require.NoError(t, err)
		require.Empty(t, resp.Digest)
		require.Equal(t, portainer.APIKeyAccessPresetReadOnly, resp.AccessPreset)
		require.Equal(t, portainer.APIKeyAccessPresetReadOnly, resp.EffectiveAccessPreset)

		updatedAPIKey, err := apiKeyService.GetAPIKey(apiKey.ID)
		require.NoError(t, err)
		require.Equal(t, portainer.APIKeyAccessPresetReadOnly, updatedAPIKey.AccessPreset)
	})

	t.Run("user can temporarily elevate access token preset", func(t *testing.T) {
		_, apiKey, err := apiKeyService.GenerateApiKey(*user, "test-temporary-elevation")
		require.NoError(t, err)

		apiKey.AccessPreset = portainer.APIKeyAccessPresetPower
		require.NoError(t, apiKeyService.UpdateAPIKey(apiKey))

		expiresAt := time.Now().UTC().Add(time.Hour).Unix()
		payload, err := json.Marshal(userAccessTokenUpdatePayload{
			AccessPreset:             portainer.APIKeyAccessPresetPower,
			TemporaryAccessPreset:    portainer.APIKeyAccessPresetManage,
			TemporaryAccessExpiresAt: expiresAt,
		})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/users/%d/tokens/%d", user.ID, apiKey.ID), bytes.NewBuffer(payload))
		testhelpers.AddTestSecurityCookie(req, jwt)

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)

		body, err := io.ReadAll(rr.Body)
		require.NoError(t, err)

		var resp apiKeyResponse
		err = json.Unmarshal(body, &resp)
		require.NoError(t, err)
		require.Equal(t, portainer.APIKeyAccessPresetPower, resp.AccessPreset)
		require.Equal(t, portainer.APIKeyAccessPresetManage, resp.TemporaryAccessPreset)
		require.Equal(t, expiresAt, resp.TemporaryAccessExpiresAt)
		require.Equal(t, portainer.APIKeyAccessPresetManage, resp.EffectiveAccessPreset)

		updatedAPIKey, err := apiKeyService.GetAPIKey(apiKey.ID)
		require.NoError(t, err)
		require.Equal(t, portainer.APIKeyAccessPresetPower, updatedAPIKey.AccessPreset)
		require.Equal(t, portainer.APIKeyAccessPresetManage, updatedAPIKey.TemporaryAccessPreset)
		require.Equal(t, expiresAt, updatedAPIKey.TemporaryAccessExpiresAt)
	})

	t.Run("user can clear temporary access token elevation", func(t *testing.T) {
		_, apiKey, err := apiKeyService.GenerateApiKey(*user, "test-clear-temporary-elevation")
		require.NoError(t, err)

		apiKey.AccessPreset = portainer.APIKeyAccessPresetPower
		apiKey.TemporaryAccessPreset = portainer.APIKeyAccessPresetManage
		apiKey.TemporaryAccessExpiresAt = time.Now().UTC().Add(time.Hour).Unix()
		require.NoError(t, apiKeyService.UpdateAPIKey(apiKey))

		payload, err := json.Marshal(userAccessTokenUpdatePayload{
			AccessPreset:             portainer.APIKeyAccessPresetPower,
			TemporaryAccessPreset:    "",
			TemporaryAccessExpiresAt: 0,
		})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/users/%d/tokens/%d", user.ID, apiKey.ID), bytes.NewBuffer(payload))
		testhelpers.AddTestSecurityCookie(req, jwt)

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)

		updatedAPIKey, err := apiKeyService.GetAPIKey(apiKey.ID)
		require.NoError(t, err)
		require.Equal(t, portainer.APIKeyAccessPresetPower, updatedAPIKey.AccessPreset)
		require.Empty(t, updatedAPIKey.TemporaryAccessPreset)
		require.Zero(t, updatedAPIKey.TemporaryAccessExpiresAt)
	})

	t.Run("user cannot update another user's access token preset", func(t *testing.T) {
		_, apiKey, err := apiKeyService.GenerateApiKey(*otherUser, "test-other-token")
		require.NoError(t, err)

		payload, err := json.Marshal(userAccessTokenUpdatePayload{AccessPreset: portainer.APIKeyAccessPresetDisabled})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/users/%d/tokens/%d", user.ID, apiKey.ID), bytes.NewBuffer(payload))
		testhelpers.AddTestSecurityCookie(req, jwt)

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		require.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("invalid access token preset is rejected", func(t *testing.T) {
		_, apiKey, err := apiKeyService.GenerateApiKey(*user, "test-invalid-token")
		require.NoError(t, err)

		payload, err := json.Marshal(userAccessTokenUpdatePayload{AccessPreset: "invalid"})
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/users/%d/tokens/%d", user.ID, apiKey.ID), bytes.NewBuffer(payload))
		testhelpers.AddTestSecurityCookie(req, jwt)

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
	})
}
