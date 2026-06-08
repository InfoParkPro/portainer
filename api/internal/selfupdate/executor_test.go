package selfupdate

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/require"
)

func TestExecutorRunsSelfUpdatePlan(t *testing.T) {
	client := &fakeDockerClient{}
	plan := Plan{
		CurrentContainerID:    "old-id",
		CurrentContainerName:  "old-portainer",
		RollbackContainerName: "portainer-rollback-20260608",
		TargetContainerName:   "portainer-current",
		TargetImage:           "ghcr.io/infoparkpro/portainer:latest",
		ContainerConfig:       &container.Config{Image: "ghcr.io/infoparkpro/portainer:latest"},
		HostConfig:            &container.HostConfig{},
		Networks: map[string]*network.EndpointSettings{
			"bridge": {NetworkID: "bridge-id"},
		},
	}

	err := ExecutePlan(context.Background(), client, plan)

	require.NoError(t, err)
	require.Equal(t, []string{
		"pull ghcr.io/infoparkpro/portainer:latest",
		"rename old-id portainer-rollback-20260608",
		"stop old-id",
		"disconnect bridge-id old-id",
		"create portainer-current",
		"start new-id",
	}, client.calls)
	require.True(t, client.pullBody.readCalled)
}

func TestExecutorCreatesContainerWithoutInitialNetworkWhenPlanHasNoNetworks(t *testing.T) {
	client := &fakeDockerClient{}
	plan := Plan{
		CurrentContainerID:    "old-id",
		RollbackContainerName: "portainer-rollback-20260608",
		TargetContainerName:   "portainer-current",
		TargetImage:           "ghcr.io/infoparkpro/portainer:latest",
		ContainerConfig:       &container.Config{Image: "ghcr.io/infoparkpro/portainer:latest"},
		HostConfig:            &container.HostConfig{},
		Networks:              map[string]*network.EndpointSettings{},
	}

	err := ExecutePlan(context.Background(), client, plan)

	require.NoError(t, err)
	require.Equal(t, map[string]*network.EndpointSettings{}, client.createdNetworks.EndpointsConfig)
}

type fakeDockerClient struct {
	calls           []string
	pullBody        *trackingReadCloser
	createdNetworks *network.NetworkingConfig
}

func (c *fakeDockerClient) ImagePull(_ context.Context, ref string, _ image.PullOptions) (io.ReadCloser, error) {
	c.calls = append(c.calls, "pull "+ref)
	c.pullBody = &trackingReadCloser{reader: strings.NewReader("pull complete")}
	return c.pullBody, nil
}

func (c *fakeDockerClient) ContainerRename(_ context.Context, containerID string, newName string) error {
	c.calls = append(c.calls, "rename "+containerID+" "+newName)
	return nil
}

func (c *fakeDockerClient) ContainerStop(_ context.Context, containerID string, _ container.StopOptions) error {
	c.calls = append(c.calls, "stop "+containerID)
	return nil
}

func (c *fakeDockerClient) NetworkDisconnect(_ context.Context, networkID string, containerID string, _ bool) error {
	c.calls = append(c.calls, "disconnect "+networkID+" "+containerID)
	return nil
}

func (c *fakeDockerClient) ContainerCreate(
	_ context.Context,
	_ *container.Config,
	_ *container.HostConfig,
	networkingConfig *network.NetworkingConfig,
	_ *ocispec.Platform,
	containerName string,
) (container.CreateResponse, error) {
	c.calls = append(c.calls, "create "+containerName)
	c.createdNetworks = networkingConfig
	return container.CreateResponse{ID: "new-id"}, nil
}

func (c *fakeDockerClient) NetworkConnect(_ context.Context, networkID string, containerID string, _ *network.EndpointSettings) error {
	c.calls = append(c.calls, "connect "+networkID+" "+containerID)
	return nil
}

func (c *fakeDockerClient) ContainerStart(_ context.Context, containerID string, _ container.StartOptions) error {
	c.calls = append(c.calls, "start "+containerID)
	return nil
}

func (c *fakeDockerClient) ContainerRemove(_ context.Context, containerID string, _ container.RemoveOptions) error {
	c.calls = append(c.calls, "remove "+containerID)
	return nil
}

type trackingReadCloser struct {
	reader     *strings.Reader
	readCalled bool
}

func (r *trackingReadCloser) Read(p []byte) (int, error) {
	r.readCalled = true
	return r.reader.Read(p)
}

func (r *trackingReadCloser) Close() error {
	return nil
}
