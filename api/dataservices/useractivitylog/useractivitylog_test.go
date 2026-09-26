package useractivitylog_test

import (
	"testing"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/datastore"
	"github.com/portainer/portainer/api/dataservices/useractivitylog"

	"github.com/stretchr/testify/require"
)

func Test_Service_LogAndLogs(t *testing.T) {
	_, store := datastore.MustNewTestStore(t, true, true)

	service, err := useractivitylog.NewService(store.Connection())
	require.NoError(t, err)

	entries := []*portainer.UserActivityLog{
		{Username: "admin", Context: "portainer", Action: "stack_create", Payload: []byte(`{"success":true}`)},
		{Username: "ci-bot", Context: "my-environment", Action: "image_pull", APIKeyID: 2, APIKey: "ci", Payload: []byte(`{"success":true,"image":"nginx"}`)},
	}

	for _, entry := range entries {
		require.NoError(t, service.Log(entry))
	}

	logs, err := service.Logs()
	require.NoError(t, err)
	require.Len(t, logs, 2)
	require.Equal(t, 1, logs[0].ID)
	require.Equal(t, 2, logs[1].ID)
	require.Equal(t, "ci-bot", logs[1].Username)
	require.Equal(t, "ci", logs[1].APIKey)
	require.JSONEq(t, `{"success":true,"image":"nginx"}`, string(logs[1].Payload))
}
