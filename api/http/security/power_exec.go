package security

import (
	"fmt"
	"path"
	"strings"

	portainer "github.com/portainer/portainer/api"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
)

// PowerAPIKeyExecCheck reports whether a container passes the Power API token
// exec safety rules. The deny reason is meant for debug logs only.
//
// Only host-side mount sources are checked: container-internal mount targets
// never expose host paths, so where a mount lands inside the container does
// not matter for host safety.
func PowerAPIKeyExecCheck(containerInfo container.InspectResponse) (allowed bool, denyReason string) {
	if containerInfo.Config == nil || containerInfo.Config.Labels[portainer.PowerAPIKeyExecLabel] != "true" {
		return false, "missing label " + portainer.PowerAPIKeyExecLabel + "=true"
	}

	hostConfig := containerInfo.HostConfig
	if hostConfig != nil {
		if hostConfig.Privileged {
			return false, "privileged container"
		}

		for _, bind := range hostConfig.Binds {
			if bindSourceIsDangerous(bind) {
				return false, fmt.Sprintf("bind %q has a dangerous host source path", bind)
			}
		}

		for _, mount := range hostConfig.Mounts {
			if mountSourceIsDangerous(mount.Source, string(mount.Type)) {
				return false, fmt.Sprintf("mount source %q of type %q is a dangerous host path", mount.Source, mount.Type)
			}
		}

		for _, capability := range hostConfig.CapAdd {
			if strings.EqualFold(capability, "SYS_ADMIN") {
				return false, "SYS_ADMIN capability"
			}
		}
	}

	for _, mount := range containerInfo.Mounts {
		if mountSourceIsDangerous(mount.Source, string(mount.Type)) {
			return false, fmt.Sprintf("mount source %q of type %q is a dangerous host path", mount.Source, mount.Type)
		}
	}

	return true, ""
}

// bindSourceIsDangerous checks only the host-side source of a short-syntax
// bind. The container-side target and the mode flags never affect host safety.
func bindSourceIsDangerous(bind string) bool {
	source, _, _ := strings.Cut(bind, ":")
	return isDangerousPowerExecPath(source)
}

// mountSourceIsDangerous checks the host-side source of a long-syntax mount.
// Volume, tmpfs, and image sources are managed by the daemon, so only bind
// sources point at a host path of the container author's choosing.
func mountSourceIsDangerous(source string, mountType string) bool {
	if source == "" {
		return false
	}

	if mountType != "" && mountType != string(mount.TypeBind) {
		return false
	}

	return isDangerousPowerExecPath(source)
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
