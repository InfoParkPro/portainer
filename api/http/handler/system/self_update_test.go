package system

import (
	"context"
	"errors"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"
)

func TestStartSelfUpdateHelperContainerRemovesContainerWhenStartFails(t *testing.T) {
	client := &fakeSelfUpdateHelperClient{startErr: errors.New("start failed")}

	err := startSelfUpdateHelperContainer(context.Background(), client, "helper-id")

	require.ErrorContains(t, err, "start failed")
	require.Equal(t, []string{"start helper-id", "remove helper-id"}, client.calls)
}

func TestStartSelfUpdateHelperContainerReportsRemovalFailure(t *testing.T) {
	client := &fakeSelfUpdateHelperClient{
		startErr:  errors.New("start failed"),
		removeErr: errors.New("remove failed"),
	}

	err := startSelfUpdateHelperContainer(context.Background(), client, "helper-id")

	require.ErrorContains(t, err, "start failed")
	require.ErrorContains(t, err, "remove failed")
}

type fakeSelfUpdateHelperClient struct {
	calls     []string
	startErr  error
	removeErr error
}

func (c *fakeSelfUpdateHelperClient) ContainerStart(_ context.Context, containerID string, _ container.StartOptions) error {
	c.calls = append(c.calls, "start "+containerID)
	return c.startErr
}

func (c *fakeSelfUpdateHelperClient) ContainerRemove(_ context.Context, containerID string, _ container.RemoveOptions) error {
	c.calls = append(c.calls, "remove "+containerID)
	return c.removeErr
}
