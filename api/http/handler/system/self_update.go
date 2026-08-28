package system

import (
	"context"
	"encoding/base64"
	"encoding/json"
	stderrors "errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/pkg/errors"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/internal/selfupdate"
	"github.com/portainer/portainer/api/logs"
	"github.com/portainer/portainer/api/platform"
	httperror "github.com/portainer/portainer/pkg/libhttp/error"
	"github.com/portainer/portainer/pkg/libhttp/request"
	"github.com/portainer/portainer/pkg/libhttp/response"
)

type selfUpdateStartPayload struct {
	TargetImage string `json:"targetImage"`
}

const selfUpdateHelperContainerName = "portainer-self-update-helper"

type selfUpdateHelperClient interface {
	ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error
	ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error
}

func (payload *selfUpdateStartPayload) Validate(r *http.Request) error {
	return nil
}

func (handler *Handler) selfUpdatePlan(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	plan, err := handler.buildSelfUpdatePlan(r.Context(), r.URL.Query().Get("targetImage"))
	if err != nil {
		return selfUpdateError(err)
	}

	return response.JSON(w, plan)
}

func (handler *Handler) selfUpdateStart(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	payload, err := request.GetPayload[selfUpdateStartPayload](r)
	if err != nil {
		return httperror.BadRequest("Invalid request payload", err)
	}

	plan, err := handler.buildSelfUpdatePlan(r.Context(), payload.TargetImage)
	if err != nil {
		return selfUpdateError(err)
	}

	if !plan.Allowed {
		return httperror.Conflict("Self-update is not available for this Portainer deployment", errors.New(plan.BlockReason))
	}

	if err := handler.startSelfUpdateHelper(r.Context(), plan); err != nil {
		return httperror.InternalServerError("Unable to start self-update helper", err)
	}

	return response.Empty(w)
}

func (handler *Handler) buildSelfUpdatePlan(ctx context.Context, targetImage string) (selfupdate.Plan, error) {
	if handler.DockerClientFactory == nil {
		return selfupdate.Plan{}, errors.New("Docker client factory is not configured")
	}

	environment, err := handler.platformService.GetLocalEnvironment()
	if err != nil {
		if errors.Is(err, platform.ErrNoLocalEnvironment) {
			return selfupdate.Plan{}, err
		}
		return selfupdate.Plan{}, errors.Wrap(err, "get local environment")
	}

	socketPath, err := localDockerSocketPath(environment)
	if err != nil {
		return selfupdate.Plan{}, err
	}
	_ = socketPath

	cli, err := handler.DockerClientFactory.CreateClient(environment, "", nil)
	if err != nil {
		return selfupdate.Plan{}, errors.Wrap(err, "create Docker client")
	}
	defer logs.CloseAndLogErr(cli)

	hostname, err := os.Hostname()
	if err != nil {
		return selfupdate.Plan{}, errors.Wrap(err, "get current hostname")
	}

	inspect, err := selfupdate.DiscoverCurrentContainer(ctx, cli, hostname)
	if err != nil {
		return selfupdate.Plan{}, err
	}

	if targetImage == "" {
		if inspect.Config != nil && inspect.Config.Image != "" {
			targetImage = inspect.Config.Image
		} else {
			targetImage = inspect.Image
		}
	}

	return selfupdate.BuildPlan(inspect, targetImage, time.Now().UTC().Format("20060102-150405")), nil
}

func (handler *Handler) startSelfUpdateHelper(ctx context.Context, plan selfupdate.Plan) error {
	environment, err := handler.platformService.GetLocalEnvironment()
	if err != nil {
		return errors.Wrap(err, "get local environment")
	}

	socketPath, err := localDockerSocketPath(environment)
	if err != nil {
		return err
	}

	cli, err := handler.DockerClientFactory.CreateClient(environment, "", nil)
	if err != nil {
		return errors.Wrap(err, "create Docker client")
	}
	defer logs.CloseAndLogErr(cli)

	encodedPlan, err := encodePlan(plan)
	if err != nil {
		return err
	}

	createResponse, err := cli.ContainerCreate(
		ctx,
		&container.Config{
			Image: plan.CurrentImage,
			Cmd:   []string{"--self-update-helper-plan", encodedPlan},
			Labels: map[string]string{
				"io.portainer.hideStack": "true",
				"io.portainer.updater":   "true",
			},
		},
		&container.HostConfig{
			AutoRemove:  true,
			Binds:       []string{socketPath + ":/var/run/docker.sock"},
			NetworkMode: "none",
		},
		&network.NetworkingConfig{},
		nil,
		selfUpdateHelperContainerName,
	)
	if err != nil {
		return errors.Wrap(err, "create self-update helper container")
	}

	return startSelfUpdateHelperContainer(ctx, cli, createResponse.ID)
}

func startSelfUpdateHelperContainer(ctx context.Context, client selfUpdateHelperClient, containerID string) error {
	if err := client.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		removeErr := client.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: true})
		return stderrors.Join(
			errors.Wrap(err, "start self-update helper container"),
			errors.Wrap(removeErr, "remove failed self-update helper container"),
		)
	}

	return nil
}

func encodePlan(plan selfupdate.Plan) (string, error) {
	payload, err := json.Marshal(plan)
	if err != nil {
		return "", errors.Wrap(err, "encode self-update plan")
	}

	return base64.StdEncoding.EncodeToString(payload), nil
}

func localDockerSocketPath(environment *portainer.Endpoint) (string, error) {
	if environment == nil {
		return "", errors.New("local environment is missing")
	}

	if !strings.HasPrefix(environment.URL, "unix://") {
		return "", errors.Errorf("self-update only supports local Docker socket environments, got %s", environment.URL)
	}

	socketPath := strings.TrimPrefix(environment.URL, "unix://")
	if socketPath == "" {
		return "", errors.New("local Docker socket path is empty")
	}

	return socketPath, nil
}

func selfUpdateError(err error) *httperror.HandlerError {
	if errors.Is(err, platform.ErrNoLocalEnvironment) {
		return httperror.NotFound("Self-update is disabled because no local Docker environment was detected.", err)
	}

	return httperror.InternalServerError("Unable to build self-update plan", err)
}
