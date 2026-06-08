package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"os"

	"github.com/docker/docker/client"
	"github.com/portainer/portainer/api/internal/selfupdate"
	"github.com/portainer/portainer/api/logs"
	"github.com/rs/zerolog/log"
)

const selfUpdateHelperPlanFlag = "--self-update-helper-plan"

func runSelfUpdateHelperMode() bool {
	encodedPlan := selfUpdateHelperPlan()
	if encodedPlan == "" {
		return false
	}

	plan, err := decodeSelfUpdatePlan(encodedPlan)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to decode self-update plan")
	}

	cli, err := client.NewClientWithOpts(
		client.WithHost("unix:///var/run/docker.sock"),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create Docker client for self-update helper")
	}
	defer logs.CloseAndLogErr(cli)

	if err := selfupdate.ExecutePlan(context.Background(), cli, plan); err != nil {
		log.Fatal().Err(err).Msg("failed to execute self-update plan")
	}

	return true
}

func selfUpdateHelperPlan() string {
	for i, arg := range os.Args {
		if arg == selfUpdateHelperPlanFlag && i+1 < len(os.Args) {
			return os.Args[i+1]
		}
	}

	return ""
}

func decodeSelfUpdatePlan(encodedPlan string) (selfupdate.Plan, error) {
	payload, err := base64.StdEncoding.DecodeString(encodedPlan)
	if err != nil {
		return selfupdate.Plan{}, err
	}

	var plan selfupdate.Plan
	if err := json.Unmarshal(payload, &plan); err != nil {
		return selfupdate.Plan{}, err
	}

	return plan, nil
}
