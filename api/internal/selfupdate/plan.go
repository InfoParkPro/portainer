package selfupdate

import (
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
)

const (
	RunModePlainContainer = "plain-container"
	RunModeCompose        = "compose"
	RunModeSwarmService   = "swarm-service"

	defaultTargetContainerName = "portainer-current"
	rollbackNamePrefix         = "portainer-rollback-"
)

type Plan struct {
	Allowed               bool                                 `json:"allowed"`
	Mode                  string                               `json:"mode"`
	BlockReason           string                               `json:"blockReason,omitempty"`
	CurrentContainerID    string                               `json:"currentContainerId"`
	CurrentContainerName  string                               `json:"currentContainerName"`
	CurrentImage          string                               `json:"currentImage"`
	TargetImage           string                               `json:"targetImage"`
	TargetContainerName   string                               `json:"targetContainerName"`
	RollbackContainerName string                               `json:"rollbackContainerName"`
	ContainerConfig       *container.Config                    `json:"containerConfig"`
	HostConfig            *container.HostConfig                `json:"hostConfig"`
	Networks              map[string]*network.EndpointSettings `json:"networks"`
}

func BuildPlan(inspect container.InspectResponse, targetImage string, suffix string) Plan {
	labels := map[string]string{}
	if inspect.Config != nil && inspect.Config.Labels != nil {
		labels = inspect.Config.Labels
	}

	mode, allowed, blockReason := detectMode(labels)

	containerConfig := copyContainerConfig(inspect.Config)
	if containerConfig != nil {
		containerConfig.Image = targetImage
	}

	return Plan{
		Allowed:               allowed,
		Mode:                  mode,
		BlockReason:           blockReason,
		CurrentContainerID:    inspect.ID,
		CurrentContainerName:  strings.TrimPrefix(inspect.Name, "/"),
		CurrentImage:          currentImage(inspect),
		TargetImage:           targetImage,
		TargetContainerName:   defaultTargetContainerName,
		RollbackContainerName: rollbackNamePrefix + suffix,
		ContainerConfig:       containerConfig,
		HostConfig:            copyHostConfig(inspect.HostConfig),
		Networks:              copyNetworks(inspect.NetworkSettings),
	}
}

func detectMode(labels map[string]string) (mode string, allowed bool, blockReason string) {
	if labels["com.docker.swarm.service.id"] != "" || labels["com.docker.swarm.service.name"] != "" {
		return RunModeSwarmService, false, "Portainer is running as a Swarm service; update it with docker service update."
	}

	if labels["com.docker.compose.project"] != "" {
		return RunModeCompose, false, "Portainer is running as a Compose container; update it with docker compose pull && docker compose up -d."
	}

	return RunModePlainContainer, true, ""
}

func currentImage(inspect container.InspectResponse) string {
	if inspect.Config != nil && inspect.Config.Image != "" {
		return inspect.Config.Image
	}

	return inspect.Image
}

func copyContainerConfig(config *container.Config) *container.Config {
	if config == nil {
		return nil
	}

	copied := *config
	copied.Hostname = ""
	copied.Env = append([]string(nil), config.Env...)
	copied.Cmd = append([]string(nil), config.Cmd...)
	copied.Entrypoint = append([]string(nil), config.Entrypoint...)
	copied.Labels = copyStringMap(config.Labels)
	copied.ExposedPorts = config.ExposedPorts

	return &copied
}

func copyHostConfig(config *container.HostConfig) *container.HostConfig {
	if config == nil {
		return nil
	}

	copied := *config
	copied.Binds = append([]string(nil), config.Binds...)
	copied.Links = append([]string(nil), config.Links...)
	copied.DNS = append([]string(nil), config.DNS...)
	copied.DNSOptions = append([]string(nil), config.DNSOptions...)
	copied.DNSSearch = append([]string(nil), config.DNSSearch...)
	copied.ExtraHosts = append([]string(nil), config.ExtraHosts...)
	copied.VolumesFrom = append([]string(nil), config.VolumesFrom...)
	copied.CapAdd = append([]string(nil), config.CapAdd...)
	copied.CapDrop = append([]string(nil), config.CapDrop...)
	copied.SecurityOpt = append([]string(nil), config.SecurityOpt...)
	copied.GroupAdd = append([]string(nil), config.GroupAdd...)
	copied.PortBindings = config.PortBindings
	copied.Mounts = append([]mount.Mount(nil), config.Mounts...)
	copied.AutoRemove = false

	return &copied
}

func copyNetworks(settings *container.NetworkSettings) map[string]*network.EndpointSettings {
	if settings == nil || settings.Networks == nil {
		return map[string]*network.EndpointSettings{}
	}

	copied := make(map[string]*network.EndpointSettings, len(settings.Networks))
	for name, endpoint := range settings.Networks {
		if endpoint == nil {
			copied[name] = nil
			continue
		}

		copied[name] = endpoint.Copy()
	}

	return copied
}

func copyStringMap(values map[string]string) map[string]string {
	if values == nil {
		return nil
	}

	copied := make(map[string]string, len(values))
	for key, value := range values {
		copied[key] = value
	}

	return copied
}
