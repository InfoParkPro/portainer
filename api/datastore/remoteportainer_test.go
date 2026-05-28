package datastore

import (
	"testing"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/dataservices"

	"github.com/stretchr/testify/require"
)

func TestRemotePortainerServiceStoresConnections(t *testing.T) {
	t.Parallel()

	_, store := MustNewTestStore(t, true, true)

	remote := &portainer.RemotePortainer{
		Name:          "standby",
		URL:           "https://standby.example",
		APIToken:      "ptr_token",
		TLSSkipVerify: true,
		CreatedAt:     1710000000,
		UpdatedAt:     1710000000,
	}

	require.NoError(t, store.RemotePortainer().Create(remote))
	require.NotZero(t, remote.ID)

	stored, err := store.RemotePortainer().Read(remote.ID)
	require.NoError(t, err)
	require.Equal(t, remote, stored)

	stored.Name = "standby-renamed"
	stored.APIToken = "ptr_new_token"
	stored.UpdatedAt = 1710000300
	require.NoError(t, store.RemotePortainer().Update(stored.ID, stored))

	require.NoError(t, store.ViewTx(func(tx dataservices.DataStoreTx) error {
		fromTx, err := tx.RemotePortainer().Read(remote.ID)
		require.NoError(t, err)
		require.Equal(t, "standby-renamed", fromTx.Name)
		require.Equal(t, "ptr_new_token", fromTx.APIToken)
		return nil
	}))
}
