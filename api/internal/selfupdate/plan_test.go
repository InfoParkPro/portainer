package selfupdate

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/require"
)

func TestBuildPlanForPlainContainer(t *testing.T) {
	inspect := container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{
			ID:   "abc123",
			Name: "/custom-portainer",
			HostConfig: &container.HostConfig{
				PortBindings: nat.PortMap{
					"9443/tcp": []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: "9443"}},
				},
				RestartPolicy: container.RestartPolicy{Name: "always"},
			},
		},
		Config: &container.Config{
			Hostname: "previous-container-id",
			Image:    "ghcr.io/infoparkpro/portainer:latest",
			Env:      []string{"TZ=UTC"},
			Labels: map[string]string{
				"existing": "label",
			},
		},
		NetworkSettings: &container.NetworkSettings{
			Networks: map[string]*network.EndpointSettings{
				"bridge": {NetworkID: "bridge-id"},
			},
		},
	}

	plan := BuildPlan(inspect, "ghcr.io/infoparkpro/portainer:latest", "20260608-150000")

	require.True(t, plan.Allowed)
	require.Equal(t, RunModePlainContainer, plan.Mode)
	require.Equal(t, "abc123", plan.CurrentContainerID)
	require.Equal(t, "custom-portainer", plan.CurrentContainerName)
	require.Equal(t, "portainer-current", plan.TargetContainerName)
	require.Equal(t, "portainer-rollback-20260608-150000", plan.RollbackContainerName)
	require.Equal(t, "ghcr.io/infoparkpro/portainer:latest", plan.ContainerConfig.Image)
	require.Empty(t, plan.ContainerConfig.Hostname)
	require.Equal(t, "0.0.0.0", plan.HostConfig.PortBindings["9443/tcp"][0].HostIP)
	require.Equal(t, "label", plan.ContainerConfig.Labels["existing"])
	require.Contains(t, plan.Networks, "bridge")
}

func TestBuildPlanDisallowsSwarmService(t *testing.T) {
	inspect := container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{
			ID:         "abc123",
			Name:       "/portainer.1.task",
			HostConfig: &container.HostConfig{},
		},
		Config: &container.Config{
			Image: "ghcr.io/infoparkpro/portainer:latest",
			Labels: map[string]string{
				"com.docker.swarm.service.id": "service-id",
			},
		},
		NetworkSettings: &container.NetworkSettings{},
	}

	plan := BuildPlan(inspect, "ghcr.io/infoparkpro/portainer:latest", "20260608-150000")

	require.False(t, plan.Allowed)
	require.Equal(t, RunModeSwarmService, plan.Mode)
	require.Contains(t, plan.BlockReason, "Swarm service")
}

func TestBuildPlanDetectsComposeContainer(t *testing.T) {
	inspect := container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{
			ID:         "abc123",
			Name:       "/compose-portainer",
			HostConfig: &container.HostConfig{},
		},
		Config: &container.Config{
			Image: "ghcr.io/infoparkpro/portainer:latest",
			Labels: map[string]string{
				"com.docker.compose.project": "portainer",
			},
		},
		NetworkSettings: &container.NetworkSettings{},
	}

	plan := BuildPlan(inspect, "ghcr.io/infoparkpro/portainer:latest", "20260608-150000")

	require.False(t, plan.Allowed)
	require.Equal(t, RunModeCompose, plan.Mode)
	require.Contains(t, plan.BlockReason, "Compose")
}
