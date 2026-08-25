package docker

import (
	"net/http"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/http/proxy/factory/utils"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/logs"
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

	if !security.PowerAPIKeyCanExecContainer(containerInfo) {
		response, err := utils.WriteAccessDeniedResponse()
		return response, true, err
	}

	return nil, false, nil
}
