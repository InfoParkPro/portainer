package selfupdate

import (
	"context"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"github.com/portainer/portainer/api/logs"
)

type DockerClient interface {
	ImagePull(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error)
	ContainerRename(ctx context.Context, containerID string, newName string) error
	ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error
	NetworkDisconnect(ctx context.Context, networkID string, containerID string, force bool) error
	ContainerCreate(
		ctx context.Context,
		config *container.Config,
		hostConfig *container.HostConfig,
		networkingConfig *network.NetworkingConfig,
		platform *ocispec.Platform,
		containerName string,
	) (container.CreateResponse, error)
	NetworkConnect(ctx context.Context, networkID string, containerID string, config *network.EndpointSettings) error
	ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error
	ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error
}

func ExecutePlan(ctx context.Context, client DockerClient, plan Plan) error {
	rc, err := client.ImagePull(ctx, plan.TargetImage, image.PullOptions{})
	if err != nil {
		return errors.Wrap(err, "pull target image")
	}
	if _, err := io.Copy(io.Discard, rc); err != nil {
		logs.CloseAndLogErr(rc)
		return errors.Wrap(err, "read image pull response")
	}
	logs.CloseAndLogErr(rc)

	if err := client.ContainerRename(ctx, plan.CurrentContainerID, plan.RollbackContainerName); err != nil {
		return errors.Wrap(err, "rename current container")
	}

	rollbackNeeded := false
	defer func() {
		if rollbackNeeded {
			_ = restoreRollback(ctx, client, plan)
		}
	}()

	if err := client.ContainerStop(ctx, plan.CurrentContainerID, container.StopOptions{}); err != nil {
		return errors.Wrap(err, "stop current container")
	}
	rollbackNeeded = true

	for _, endpoint := range plan.Networks {
		if endpoint == nil || endpoint.NetworkID == "" {
			continue
		}

		if err := client.NetworkDisconnect(ctx, endpoint.NetworkID, plan.CurrentContainerID, true); err != nil {
			return errors.Wrap(err, "disconnect rollback container network")
		}
	}

	initialNetworkName, initialNetwork := firstNetwork(plan.Networks)
	networkingConfig := &network.NetworkingConfig{EndpointsConfig: map[string]*network.EndpointSettings{}}
	if initialNetworkName != "" {
		networkingConfig.EndpointsConfig[initialNetworkName] = initialNetwork
	}

	createResponse, err := client.ContainerCreate(
		ctx,
		plan.ContainerConfig,
		plan.HostConfig,
		networkingConfig,
		nil,
		plan.TargetContainerName,
	)
	if err != nil {
		return errors.Wrap(err, "create target container")
	}

	for name, endpoint := range plan.Networks {
		if name == initialNetworkName || endpoint == nil || endpoint.NetworkID == "" {
			continue
		}

		if err := client.NetworkConnect(ctx, endpoint.NetworkID, createResponse.ID, endpoint); err != nil {
			_ = client.ContainerRemove(ctx, createResponse.ID, container.RemoveOptions{Force: true})
			return errors.Wrap(err, "connect target container network")
		}
	}

	if err := client.ContainerStart(ctx, createResponse.ID, container.StartOptions{}); err != nil {
		_ = client.ContainerRemove(ctx, createResponse.ID, container.RemoveOptions{Force: true})
		return errors.Wrap(err, "start target container")
	}

	rollbackNeeded = false
	return nil
}

func restoreRollback(ctx context.Context, client DockerClient, plan Plan) error {
	for _, endpoint := range plan.Networks {
		if endpoint == nil || endpoint.NetworkID == "" {
			continue
		}

		_ = client.NetworkConnect(ctx, endpoint.NetworkID, plan.CurrentContainerID, endpoint)
	}

	return client.ContainerStart(ctx, plan.CurrentContainerID, container.StartOptions{})
}

func firstNetwork(networks map[string]*network.EndpointSettings) (string, *network.EndpointSettings) {
	for name, endpoint := range networks {
		return name, endpoint
	}

	return "", nil
}
