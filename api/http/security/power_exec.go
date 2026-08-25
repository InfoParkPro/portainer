package security

import (
	"path"
	"strings"

	portainer "github.com/portainer/portainer/api"

	"github.com/docker/docker/api/types/container"
)

func PowerAPIKeyCanExecContainer(containerInfo container.InspectResponse) bool {
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
