package selfupdate

import (
	"context"
	stderrors "errors"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/errdefs"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"github.com/portainer/portainer/api/logs"
)

type DockerClient interface {
	ImagePull(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error)
	ContainerInspect(ctx context.Context, containerID string) (container.InspectResponse, error)
	ContainerUpdate(ctx context.Context, containerID string, updateConfig container.UpdateConfig) (container.UpdateResponse, error)
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

func ExecutePlan(ctx context.Context, client DockerClient, plan Plan) (returnErr error) {
	targetContainer, err := client.ContainerInspect(ctx, plan.TargetContainerName)
	if err == nil && targetContainer.ID != plan.CurrentContainerID {
		return errors.Errorf("target name %s already belongs to container %s", plan.TargetContainerName, targetContainer.ID)
	}
	if err != nil && !errdefs.IsNotFound(err) {
		return errors.Wrap(err, "inspect target container name")
	}

	rc, err := client.ImagePull(ctx, plan.TargetImage, image.PullOptions{})
	if err != nil {
		return errors.Wrap(err, "pull target image")
	}
	if _, err := io.Copy(io.Discard, rc); err != nil {
		logs.CloseAndLogErr(rc)
		return errors.Wrap(err, "read image pull response")
	}
	logs.CloseAndLogErr(rc)

	if _, err := client.ContainerUpdate(ctx, plan.CurrentContainerID, container.UpdateConfig{
		RestartPolicy: container.RestartPolicy{Name: container.RestartPolicyDisabled},
	}); err != nil {
		return errors.Wrap(err, "disable rollback container restart policy")
	}
	restartPolicyDisabled := true
	renamed := false
	defer func() {
		if returnErr == nil || !restartPolicyDisabled {
			return
		}

		var rollbackErr error
		if renamed {
			rollbackErr = restoreRollback(ctx, client, plan)
		} else {
			rollbackErr = restoreRestartPolicy(ctx, client, plan)
		}
		if rollbackErr != nil {
			returnErr = stderrors.Join(returnErr, errors.Wrap(rollbackErr, "restore current container after failed self-update"))
		}
	}()

	if err := client.ContainerRename(ctx, plan.CurrentContainerID, plan.RollbackContainerName); err != nil {
		return errors.Wrap(err, "rename current container")
	}
	renamed = true

	if err := client.ContainerStop(ctx, plan.CurrentContainerID, container.StopOptions{}); err != nil {
		return errors.Wrap(err, "stop current container")
	}

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
			removeErr := client.ContainerRemove(ctx, createResponse.ID, container.RemoveOptions{Force: true})
			return stderrors.Join(
				errors.Wrap(err, "connect target container network"),
				errors.Wrap(removeErr, "remove failed target container"),
			)
		}
	}

	if err := client.ContainerStart(ctx, createResponse.ID, container.StartOptions{}); err != nil {
		removeErr := client.ContainerRemove(ctx, createResponse.ID, container.RemoveOptions{Force: true})
		return stderrors.Join(
			errors.Wrap(err, "start target container"),
			errors.Wrap(removeErr, "remove failed target container"),
		)
	}

	restartPolicyDisabled = false
	return nil
}

func restoreRollback(ctx context.Context, client DockerClient, plan Plan) error {
	var restoreErr error
	for _, endpoint := range plan.Networks {
		if endpoint == nil || endpoint.NetworkID == "" {
			continue
		}

		if err := client.NetworkConnect(ctx, endpoint.NetworkID, plan.CurrentContainerID, endpoint); err != nil {
			restoreErr = stderrors.Join(restoreErr, errors.Wrap(err, "reconnect rollback container network"))
		}
	}

	if err := restoreRestartPolicy(ctx, client, plan); err != nil {
		restoreErr = stderrors.Join(restoreErr, err)
	}

	if plan.CurrentContainerName != "" && plan.CurrentContainerName != plan.RollbackContainerName {
		if err := client.ContainerRename(ctx, plan.CurrentContainerID, plan.CurrentContainerName); err != nil {
			restoreErr = stderrors.Join(restoreErr, errors.Wrap(err, "restore rollback container name"))
		}
	}

	inspect, err := client.ContainerInspect(ctx, plan.CurrentContainerID)
	if err != nil {
		restoreErr = stderrors.Join(restoreErr, errors.Wrap(err, "inspect rollback container state"))
	} else if inspect.State != nil && inspect.State.Running {
		return restoreErr
	}

	if err := client.ContainerStart(ctx, plan.CurrentContainerID, container.StartOptions{}); err != nil && !errdefs.IsNotModified(err) {
		restoreErr = stderrors.Join(restoreErr, errors.Wrap(err, "start rollback container"))
	}

	return restoreErr
}

func restoreRestartPolicy(ctx context.Context, client DockerClient, plan Plan) error {
	if plan.HostConfig == nil {
		return nil
	}

	_, err := client.ContainerUpdate(ctx, plan.CurrentContainerID, container.UpdateConfig{
		RestartPolicy: plan.HostConfig.RestartPolicy,
	})
	return errors.Wrap(err, "restore rollback container restart policy")
}

func firstNetwork(networks map[string]*network.EndpointSettings) (string, *network.EndpointSettings) {
	for name, endpoint := range networks {
		return name, endpoint
	}

	return "", nil
}
