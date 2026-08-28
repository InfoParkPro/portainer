package selfupdate

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/errdefs"
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
		"inspect portainer-current",
		"pull ghcr.io/infoparkpro/portainer:latest",
		"update-restart old-id no",
		"rename old-id portainer-rollback-20260608",
		"stop old-id",
		"disconnect bridge-id old-id",
		"create portainer-current",
		"start new-id",
	}, client.calls)
	require.True(t, client.pullBody.readCalled)
}

func TestExecutorStopsBeforeMutationWhenTargetNameBelongsToAnotherContainer(t *testing.T) {
	client := &fakeDockerClient{targetContainerID: "other-id"}
	plan := Plan{
		CurrentContainerID:  "old-id",
		TargetContainerName: "portainer-current",
		TargetImage:         "ghcr.io/infoparkpro/portainer:latest",
	}

	err := ExecutePlan(context.Background(), client, plan)

	require.ErrorContains(t, err, "already belongs to container other-id")
	require.Equal(t, []string{"inspect portainer-current"}, client.calls)
}

func TestExecutorRestoresOriginalRestartPolicyWhenRollbackIsActivated(t *testing.T) {
	client := &fakeDockerClient{createErr: errors.New("create failed")}
	plan := Plan{
		CurrentContainerID:    "old-id",
		CurrentContainerName:  "old-portainer",
		RollbackContainerName: "portainer-rollback-20260608",
		TargetContainerName:   "portainer-current",
		TargetImage:           "ghcr.io/infoparkpro/portainer:latest",
		ContainerConfig:       &container.Config{Image: "ghcr.io/infoparkpro/portainer:latest"},
		HostConfig: &container.HostConfig{
			RestartPolicy: container.RestartPolicy{Name: container.RestartPolicyAlways},
		},
		Networks: map[string]*network.EndpointSettings{
			"bridge": {NetworkID: "bridge-id"},
		},
	}

	err := ExecutePlan(context.Background(), client, plan)

	require.ErrorContains(t, err, "create target container")
	require.Equal(t, []string{
		"inspect portainer-current",
		"pull ghcr.io/infoparkpro/portainer:latest",
		"update-restart old-id no",
		"rename old-id portainer-rollback-20260608",
		"stop old-id",
		"disconnect bridge-id old-id",
		"create portainer-current",
		"connect bridge-id old-id",
		"update-restart old-id always",
		"rename old-id old-portainer",
		"inspect old-id",
		"start old-id",
	}, client.calls)
}

func TestExecutorRestoresRestartPolicyWhenRenameFails(t *testing.T) {
	client := &fakeDockerClient{renameErr: errors.New("rename failed")}
	plan := Plan{
		CurrentContainerID:    "old-id",
		CurrentContainerName:  "old-portainer",
		RollbackContainerName: "portainer-rollback-20260608",
		TargetContainerName:   "portainer-current",
		TargetImage:           "ghcr.io/infoparkpro/portainer:latest",
		ContainerConfig:       &container.Config{Image: "ghcr.io/infoparkpro/portainer:latest"},
		HostConfig: &container.HostConfig{
			RestartPolicy: container.RestartPolicy{Name: container.RestartPolicyAlways},
		},
	}

	err := ExecutePlan(context.Background(), client, plan)

	require.ErrorContains(t, err, "rename current container")
	require.Equal(t, []string{
		"inspect portainer-current",
		"pull ghcr.io/infoparkpro/portainer:latest",
		"update-restart old-id no",
		"rename old-id portainer-rollback-20260608",
		"update-restart old-id always",
	}, client.calls)
}

func TestExecutorStillStartsRollbackWhenRestartPolicyRestoreFails(t *testing.T) {
	client := &fakeDockerClient{
		createErr:         errors.New("create failed"),
		rollbackUpdateErr: errors.New("policy restore failed"),
	}
	plan := Plan{
		CurrentContainerID:    "old-id",
		CurrentContainerName:  "old-portainer",
		RollbackContainerName: "portainer-rollback-20260608",
		TargetContainerName:   "portainer-current",
		TargetImage:           "ghcr.io/infoparkpro/portainer:latest",
		ContainerConfig:       &container.Config{Image: "ghcr.io/infoparkpro/portainer:latest"},
		HostConfig: &container.HostConfig{
			RestartPolicy: container.RestartPolicy{Name: container.RestartPolicyAlways},
		},
	}

	err := ExecutePlan(context.Background(), client, plan)

	require.ErrorContains(t, err, "policy restore failed")
	require.Contains(t, client.calls, "start old-id")
}

func TestExecutorReportsTargetRemovalFailure(t *testing.T) {
	client := &fakeDockerClient{
		targetStartErr: errors.New("target start failed"),
		removeErr:      errors.New("target remove failed"),
	}
	plan := Plan{
		CurrentContainerID:    "old-id",
		CurrentContainerName:  "old-portainer",
		RollbackContainerName: "portainer-rollback-20260608",
		TargetContainerName:   "portainer-current",
		TargetImage:           "ghcr.io/infoparkpro/portainer:latest",
		ContainerConfig:       &container.Config{Image: "ghcr.io/infoparkpro/portainer:latest"},
		HostConfig: &container.HostConfig{
			RestartPolicy: container.RestartPolicy{Name: container.RestartPolicyAlways},
		},
	}

	err := ExecutePlan(context.Background(), client, plan)

	require.ErrorContains(t, err, "target start failed")
	require.ErrorContains(t, err, "target remove failed")
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
	calls             []string
	pullBody          *trackingReadCloser
	createdNetworks   *network.NetworkingConfig
	targetContainerID string
	createErr         error
	renameErr         error
	rollbackUpdateErr error
	targetStartErr    error
	removeErr         error
	updateCalls       int
}

func (c *fakeDockerClient) ContainerInspect(_ context.Context, containerID string) (container.InspectResponse, error) {
	c.calls = append(c.calls, "inspect "+containerID)
	if containerID != "portainer-current" {
		return container.InspectResponse{ContainerJSONBase: &container.ContainerJSONBase{
			ID:    containerID,
			State: &container.State{Running: false},
		}}, nil
	}
	if c.targetContainerID == "" {
		return container.InspectResponse{}, errdefs.NotFound(errors.New("not found"))
	}

	return container.InspectResponse{ContainerJSONBase: &container.ContainerJSONBase{ID: c.targetContainerID}}, nil
}

func (c *fakeDockerClient) ContainerUpdate(_ context.Context, containerID string, updateConfig container.UpdateConfig) (container.UpdateResponse, error) {
	c.calls = append(c.calls, "update-restart "+containerID+" "+string(updateConfig.RestartPolicy.Name))
	c.updateCalls++
	if c.updateCalls > 1 && c.rollbackUpdateErr != nil {
		return container.UpdateResponse{}, c.rollbackUpdateErr
	}
	return container.UpdateResponse{}, nil
}

func (c *fakeDockerClient) ImagePull(_ context.Context, ref string, _ image.PullOptions) (io.ReadCloser, error) {
	c.calls = append(c.calls, "pull "+ref)
	c.pullBody = &trackingReadCloser{reader: strings.NewReader("pull complete")}
	return c.pullBody, nil
}

func (c *fakeDockerClient) ContainerRename(_ context.Context, containerID string, newName string) error {
	c.calls = append(c.calls, "rename "+containerID+" "+newName)
	return c.renameErr
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
	if c.createErr != nil {
		return container.CreateResponse{}, c.createErr
	}
	return container.CreateResponse{ID: "new-id"}, nil
}

func (c *fakeDockerClient) NetworkConnect(_ context.Context, networkID string, containerID string, _ *network.EndpointSettings) error {
	c.calls = append(c.calls, "connect "+networkID+" "+containerID)
	return nil
}

func (c *fakeDockerClient) ContainerStart(_ context.Context, containerID string, _ container.StartOptions) error {
	c.calls = append(c.calls, "start "+containerID)
	if containerID == "new-id" {
		return c.targetStartErr
	}
	return nil
}

func (c *fakeDockerClient) ContainerRemove(_ context.Context, containerID string, _ container.RemoveOptions) error {
	c.calls = append(c.calls, "remove "+containerID)
	return c.removeErr
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
