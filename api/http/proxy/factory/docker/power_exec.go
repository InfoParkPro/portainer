package docker

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/http/proxy/factory/utils"
	"github.com/portainer/portainer/api/http/security"
	"github.com/portainer/portainer/api/logs"
)

func (transport *Transport) restrictPowerAPIKeyExecCreate(request *http.Request, containerID string) (*http.Response, bool, error) {
	tokenData, err := security.RetrieveTokenData(request)
	if err != nil {
		return nil, true, err
	}
	if tokenData.APIKeyID == 0 || tokenData.APIKeyAccessPreset != portainer.APIKeyAccessPresetPower {
		return nil, false, nil
	}

	allowed, err := powerExecCreateRequestAllowed(request)
	if err != nil {
		return nil, true, err
	}
	if !allowed {
		response, err := utils.WriteAccessDeniedResponse()
		return response, true, err
	}

	return transport.restrictPowerAPIKeyExecContainer(request, containerID)
}

func powerExecCreateRequestAllowed(request *http.Request) (bool, error) {
	if request.Body == nil {
		return true, nil
	}

	payload, err := io.ReadAll(request.Body)
	if err != nil {
		return false, err
	}
	request.Body = io.NopCloser(bytes.NewReader(payload))
	if len(payload) == 0 {
		return true, nil
	}

	var config struct {
		Privileged bool `json:"Privileged"`
	}
	if err := json.Unmarshal(payload, &config); err != nil {
		return false, nil
	}

	return !config.Privileged, nil
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

	response, err := transport.executeDockerRequest(request)
	return response, true, err
}
