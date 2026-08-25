package docker

import (
	"net/http"
	"path"
	"strings"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/http/proxy/factory/utils"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/logs"

	"github.com/docker/docker/api/types/container"
)

func (transport *Transport) restrictPowerAPIKeyExecCreate(request *http.Request, containerID string) (*http.Response, bool, error) {
	return transport.restrictPowerAPIKeyExecContainer(request, containerID)
}

func (transport *Transport) restrictPowerAPIKeyExecContainer(request *http.Request, containerID string) (*http.Response, bool, error) {
	tokenData, err := security.RetrieveTokenData(request)
	if err != nil {
		return nil, true, err
	}

	if tokenData.APIKeyID == 0 || tokenData.APIKeyAccessPreset != portainer.APIKeyAccessPresetPower {
		return nil, false, nil
	}

	client, err := transport.dockerClientFactory.CreateClient(transport.endpoint, request.Header.Get(portainer.PortainerAgentTargetHeader), nil)
	if err != nil {
		return nil, true, err
	}
	defer logs.CloseAndLogErr(client)

	containerInfo, err := client.ContainerInspect(request.Context(), containerID)
	if err != nil {
		return nil, true, err
	}

	if !powerAPIKeyCanExecContainer(containerInfo) {
		response, err := utils.WriteAccessDeniedResponse()
		return response, true, err
	}

	return nil, false, nil
}

func powerAPIKeyCanExecContainer(containerInfo container.InspectResponse) bool {
	if containerInfo.Config == nil || containerInfo.Config.Labels[portainer.PowerAPIKeyExecLabel] != "true" {
		return false
	}

	hostConfig := containerInfo.HostConfig
	if hostConfig != nil {
		if hostConfig.Privileged {
			return false
		}

		for _, bind := range hostConfig.Binds {
			if bindMountHasDangerousPath(bind) {
				return false
			}
		}

		for _, mount := range hostConfig.Mounts {
			if isDangerousPowerExecPath(mount.Source) || isDangerousPowerExecPath(mount.Target) {
				return false
			}
		}

		for _, capability := range hostConfig.CapAdd {
			if strings.EqualFold(capability, "SYS_ADMIN") {
				return false
			}
		}
	}

	for _, mount := range containerInfo.Mounts {
		if isDangerousPowerExecPath(mount.Source) || isDangerousPowerExecPath(mount.Destination) {
			return false
		}
	}

	return true
}

func bindMountHasDangerousPath(bind string) bool {
	parts := strings.Split(bind, ":")
	for _, part := range parts {
		if isDangerousPowerExecPath(part) {
			return true
		}
	}

	return false
}

func isDangerousPowerExecPath(rawPath string) bool {
	if rawPath == "" {
		return false
	}

	cleanPath := path.Clean(rawPath)

	if cleanPath == "/var/run/docker.sock" || cleanPath == "/run/docker.sock" {
		return true
	}

	dangerousPrefixes := []string{
		"/",
		"/dev",
		"/etc/docker",
		"/proc",
		"/root",
		"/run",
		"/sys",
		"/var/lib/docker",
		"/var/run",
	}

	for _, prefix := range dangerousPrefixes {
		if cleanPath == prefix || strings.HasPrefix(cleanPath, prefix+"/") {
			return true
		}
	}

	return false
}
