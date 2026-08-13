package stacks

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/datastore"
	"github.com/portainer/portainer/api/filesystem"
	"github.com/portainer/portainer/api/internal/testhelpers"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_webhookInvoke(t *testing.T) {
	t.Parallel()
	_, store := datastore.MustNewTestStore(t, false, true)
	require.NoError(t, store.User().Create(&portainer.User{
		ID:       1,
		Username: "admin",
		Role:     portainer.AdministratorRole,
	}))
	require.NoError(t, store.Endpoint().Create(&portainer.Endpoint{ID: 1}))

	webhookID := newGuidString(t)
	err := store.StackService.Create(&portainer.Stack{
		ID:         1,
		Name:       "test-stack",
		Type:       portainer.DockerComposeStack,
		EndpointID: 1,
		CreatedBy:  "admin",
		AutoUpdate: &portainer.AutoUpdateSettings{
			Webhook: webhookID,
		},
	})
	require.NoError(t, err)

	h := NewHandler(testhelpers.NewTestRequestBouncer(), nil)
	h.DataStore = store
	h.FileService, err = filesystem.NewService(t.TempDir(), "")
	require.NoError(t, err)
	h.StackDeployer = testhelpers.NewTestStackDeployer()

	t.Run("invalid uuid results in http.StatusBadRequest", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := newRequest("notuuid")
		h.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("registered webhook ID in http.StatusNoContent", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := newRequest(webhookID)
		h.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("unregistered webhook ID in http.StatusNotFound", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := newRequest(newGuidString(t))
		h.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestHandler_webhookInvokeRedeploysFileBasedStack(t *testing.T) {
	t.Parallel()

	_, store := datastore.MustNewTestStore(t, false, true)
	require.NoError(t, store.User().Create(&portainer.User{
		ID:       1,
		Username: "admin",
		Role:     portainer.AdministratorRole,
	}))
	require.NoError(t, store.Endpoint().Create(&portainer.Endpoint{ID: 1}))

	webhookID := newGuidString(t)
	require.NoError(t, store.StackService.Create(&portainer.Stack{
		ID:         1,
		Name:       "test-stack",
		Type:       portainer.DockerComposeStack,
		EndpointID: 1,
		CreatedBy:  "admin",
		AutoUpdate: &portainer.AutoUpdateSettings{
			Webhook: webhookID,
		},
	}))

	fileService, err := filesystem.NewService(t.TempDir(), "")
	require.NoError(t, err)
	deployer := testhelpers.NewTestStackDeployer()

	h := NewHandler(testhelpers.NewTestRequestBouncer(), nil)
	h.DataStore = store
	h.FileService = fileService
	h.StackDeployer = deployer

	w := httptest.NewRecorder()
	req := newRequest(webhookID)
	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	require.Eventually(t, func() bool {
		return deployer.DeployComposeCallCount == 1
	}, 5*time.Second, 10*time.Millisecond)
}

func TestHandler_webhookInvokeSkipsCooldown(t *testing.T) {
	t.Parallel()

	_, store := datastore.MustNewTestStore(t, false, true)
	require.NoError(t, store.User().Create(&portainer.User{
		ID:       1,
		Username: "admin",
		Role:     portainer.AdministratorRole,
	}))
	require.NoError(t, store.Endpoint().Create(&portainer.Endpoint{ID: 1}))

	webhookID := newGuidString(t)
	require.NoError(t, store.StackService.Create(&portainer.Stack{
		ID:         1,
		Name:       "test-stack",
		Type:       portainer.DockerComposeStack,
		EndpointID: 1,
		CreatedBy:  "admin",
		AutoUpdate: &portainer.AutoUpdateSettings{
			Webhook: webhookID,
		},
	}))

	fileService, err := filesystem.NewService(t.TempDir(), "")
	require.NoError(t, err)
	deployer := testhelpers.NewTestStackDeployer()

	h := NewHandler(testhelpers.NewTestRequestBouncer(), nil)
	h.DataStore = store
	h.FileService = fileService
	h.StackDeployer = deployer

	firstResponse := httptest.NewRecorder()
	h.ServeHTTP(firstResponse, newRequest(webhookID))
	assert.Equal(t, http.StatusNoContent, firstResponse.Code)
	require.Eventually(t, func() bool {
		return deployer.DeployComposeCallCount == 1
	}, 5*time.Second, 10*time.Millisecond)

	stored, err := store.StackService.Read(1)
	require.NoError(t, err)
	require.NotNil(t, stored.AutoUpdate)
	require.NotZero(t, stored.AutoUpdate.LastWebhookInvoke)

	secondResponse := httptest.NewRecorder()
	h.ServeHTTP(secondResponse, newRequest(webhookID))

	assert.Equal(t, http.StatusNoContent, secondResponse.Code)
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 1, deployer.DeployComposeCallCount)
}

func TestHandler_webhookInvokeAcceptsAfterCooldown(t *testing.T) {
	t.Parallel()

	_, store := datastore.MustNewTestStore(t, false, true)
	require.NoError(t, store.User().Create(&portainer.User{
		ID:       1,
		Username: "admin",
		Role:     portainer.AdministratorRole,
	}))
	require.NoError(t, store.Endpoint().Create(&portainer.Endpoint{ID: 1}))

	webhookID := newGuidString(t)
	require.NoError(t, store.StackService.Create(&portainer.Stack{
		ID:         1,
		Name:       "test-stack",
		Type:       portainer.DockerComposeStack,
		EndpointID: 1,
		CreatedBy:  "admin",
		AutoUpdate: &portainer.AutoUpdateSettings{
			Webhook:           webhookID,
			LastWebhookInvoke: time.Now().Add(-stackWebhookCooldown).Add(-time.Second).Unix(),
		},
	}))

	fileService, err := filesystem.NewService(t.TempDir(), "")
	require.NoError(t, err)
	deployer := testhelpers.NewTestStackDeployer()

	h := NewHandler(testhelpers.NewTestRequestBouncer(), nil)
	h.DataStore = store
	h.FileService = fileService
	h.StackDeployer = deployer

	w := httptest.NewRecorder()
	h.ServeHTTP(w, newRequest(webhookID))

	assert.Equal(t, http.StatusNoContent, w.Code)
	require.Eventually(t, func() bool {
		return deployer.DeployComposeCallCount == 1
	}, 5*time.Second, 10*time.Millisecond)
}

func newGuidString(t *testing.T) string {
	uuid, err := uuid.NewRandom()
	require.NoError(t, err)

	return uuid.String()
}

func newRequest(webhookID string) *http.Request {
	return httptest.NewRequest(http.MethodPost, "/stacks/webhooks/"+webhookID, nil)
}
