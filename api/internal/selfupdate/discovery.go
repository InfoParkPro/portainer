package selfupdate

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/pkg/errors"
)

type DiscoveryClient interface {
	ContainerList(ctx context.Context, options container.ListOptions) ([]container.Summary, error)
	ContainerInspect(ctx context.Context, containerID string) (container.InspectResponse, error)
}

func DiscoverCurrentContainer(ctx context.Context, client DiscoveryClient, hostname string) (container.InspectResponse, error) {
	containers, err := client.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return container.InspectResponse{}, errors.Wrap(err, "list running containers")
	}

	var namedContainers []container.Summary
	for _, candidate := range containers {
		if hasContainerName(candidate.Names, defaultTargetContainerName) {
			namedContainers = append(namedContainers, candidate)
		}
	}

	if len(namedContainers) > 1 {
		return container.InspectResponse{}, errors.New("multiple running containers named portainer-current")
	}

	containerID := hostname
	if len(namedContainers) == 1 {
		containerID = namedContainers[0].ID
	}

	inspect, err := client.ContainerInspect(ctx, containerID)
	if err != nil {
		return container.InspectResponse{}, errors.Wrap(err, "inspect current Portainer container")
	}
	if inspect.State == nil || !inspect.State.Running {
		return container.InspectResponse{}, errors.Errorf("selected Portainer container %s is not running", inspect.ID)
	}

	return inspect, nil
}

func hasContainerName(names []string, expected string) bool {
	for _, name := range names {
		if strings.TrimPrefix(name, "/") == expected {
			return true
		}
	}

	return false
}
