package stackbuilders

import (
	"testing"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/datastore"
	"github.com/portainer/portainer/api/filesystem"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/internal/testhelpers"

	"github.com/stretchr/testify/require"
)

func TestComposeStackFileBuilderStoresWebhook(t *testing.T) {
	t.Parallel()

	builder := newComposeFileStackBuilder(t)
	payload := &StackPayload{
		Name:             "test-stack",
		StackFileContent: []byte("services:\n  app:\n    image: nginx"),
		Webhook:          "8dce8c2f-9ca1-482b-ad20-271e86536ada",
	}

	err := builder.prepare(t.Context(), payload, 1)

	require.NoError(t, err)
	require.NotNil(t, builder.stack.AutoUpdate)
	require.Equal(t, payload.Webhook, builder.stack.AutoUpdate.Webhook)
}

func TestSwarmStackFileBuilderStoresWebhook(t *testing.T) {
	t.Parallel()

	builder := newSwarmFileStackBuilder(t)
	payload := &StackPayload{
		Name:             "test-stack",
		SwarmID:          "swarm-id",
		StackFileContent: []byte("services:\n  app:\n    image: nginx"),
		Webhook:          "8dce8c2f-9ca1-482b-ad20-271e86536ada",
	}

	err := builder.prepare(t.Context(), payload, 1)

	require.NoError(t, err)
	require.NotNil(t, builder.stack.AutoUpdate)
	require.Equal(t, payload.Webhook, builder.stack.AutoUpdate.Webhook)
}

func newComposeFileStackBuilder(t *testing.T) *ComposeStackFileBuilder {
	t.Helper()

	_, store := datastore.MustNewTestStore(t, false, true)
	require.NoError(t, store.User().Create(&portainer.User{ID: 1, Username: "admin"}))

	fileService, err := filesystem.NewService(t.TempDir(), "")
	require.NoError(t, err)

	builder := CreateComposeStackFileBuilder(
		&security.RestrictedRequestContext{UserID: 1, IsAdmin: true},
		store,
		fileService,
		testhelpers.NewTestStackDeployer(),
	)
	builder.setGeneralInfo(&StackPayload{}, &portainer.Endpoint{ID: 1})

	return builder
}

func newSwarmFileStackBuilder(t *testing.T) *SwarmStackFileBuilder {
	t.Helper()

	_, store := datastore.MustNewTestStore(t, false, true)
	require.NoError(t, store.User().Create(&portainer.User{ID: 1, Username: "admin"}))

	fileService, err := filesystem.NewService(t.TempDir(), "")
	require.NoError(t, err)

	builder := CreateSwarmStackFileBuilder(
		&security.RestrictedRequestContext{UserID: 1, IsAdmin: true},
		store,
		fileService,
		testhelpers.NewTestStackDeployer(),
	)
	builder.setGeneralInfo(&StackPayload{}, &portainer.Endpoint{ID: 1})

	return builder
}
