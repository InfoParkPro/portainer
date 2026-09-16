package security

import (
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	portainer "github.com/portainer/portainer/api"
	"github.com/stretchr/testify/assert"
)

func powerExecTestContainer(labels map[string]string, hostConfig *container.HostConfig, mounts []container.MountPoint) container.InspectResponse {
	return container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{HostConfig: hostConfig},
		Config:            &container.Config{Labels: labels},
		Mounts:            mounts,
	}
}

func powerExecLabel() map[string]string {
	return map[string]string{portainer.PowerAPIKeyExecLabel: "true"}
}

func Test_PowerAPIKeyExecCheck_Label(t *testing.T) {
	allowed, reason := PowerAPIKeyExecCheck(powerExecTestContainer(nil, nil, nil))
	assert.False(t, allowed)
	assert.NotEmpty(t, reason)

	labels := map[string]string{portainer.PowerAPIKeyExecLabel: "false"}
	allowed, _ = PowerAPIKeyExecCheck(powerExecTestContainer(labels, nil, nil))
	assert.False(t, allowed)
}

func Test_PowerAPIKeyExecCheck_BindSourcesOnly(t *testing.T) {
	tests := []struct {
		name    string
		binds   []string
		allowed bool
	}{
		{
			name:    "plain app-data binds with container-internal targets",
			binds:   []string{"/apps/9router/data:/app/data", "/apps/9router/usage:/root/.9router"},
			allowed: true,
		},
		{
			name:    "bind options and read-only flags are ignored",
			binds:   []string{"/apps/data:/data:ro,z"},
			allowed: true,
		},
		{
			name:    "named volume short syntax",
			binds:   []string{"9router-data:/app/data"},
			allowed: true,
		},
		{
			name:    "container-internal target alone is not dangerous",
			binds:   []string{"/apps/data:/var/run/docker.sock"},
			allowed: true,
		},
		{
			name:    "docker.sock bind source",
			binds:   []string{"/var/run/docker.sock:/var/run/docker.sock"},
			allowed: false,
		},
		{
			name:    "host root bind source",
			binds:   []string{"/:/host"},
			allowed: false,
		},
		{
			name:    "docker lib bind source",
			binds:   []string{"/var/lib/docker:/docker-lib"},
			allowed: false,
		},
		{
			name:    "run bind source",
			binds:   []string{"/run/docker.sock:/sock"},
			allowed: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			info := powerExecTestContainer(powerExecLabel(), &container.HostConfig{Binds: test.binds}, nil)
			allowed, reason := PowerAPIKeyExecCheck(info)
			assert.Equal(t, test.allowed, allowed)
			if !test.allowed {
				assert.NotEmpty(t, reason)
			}
		})
	}
}

func Test_PowerAPIKeyExecCheck_LongSyntaxMountSources(t *testing.T) {
	tests := []struct {
		name    string
		mounts  []mount.Mount
		allowed bool
	}{
		{
			name:    "bind source under apps",
			mounts:  []mount.Mount{{Type: mount.TypeBind, Source: "/apps/9router/data", Target: "/app/data"}},
			allowed: true,
		},
		{
			name:    "volume source is daemon-managed",
			mounts:  []mount.Mount{{Type: mount.TypeVolume, Source: "9router-data", Target: "/app/data"}},
			allowed: true,
		},
		{
			name:    "tmpfs target under proc-like path",
			mounts:  []mount.Mount{{Type: mount.TypeTmpfs, Target: "/proc/dat"}},
			allowed: true,
		},
		{
			name:    "bind docker.sock source",
			mounts:  []mount.Mount{{Type: mount.TypeBind, Source: "/var/run/docker.sock", Target: "/var/run/docker.sock"}},
			allowed: false,
		},
		{
			name:    "unset type with absolute source is treated as bind",
			mounts:  []mount.Mount{{Source: "/etc/docker"}},
			allowed: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			info := powerExecTestContainer(powerExecLabel(), &container.HostConfig{Mounts: test.mounts}, nil)
			allowed, reason := PowerAPIKeyExecCheck(info)
			assert.Equal(t, test.allowed, allowed)
			if !test.allowed {
				assert.NotEmpty(t, reason)
			}
		})
	}
}

func Test_PowerAPIKeyExecCheck_ResolvedMounts(t *testing.T) {
	info := powerExecTestContainer(powerExecLabel(), nil, []container.MountPoint{
		{Type: mount.TypeVolume, Source: "/var/lib/docker/volumes/9router-data/_data", Destination: "/app/data"},
		{Type: mount.TypeBind, Source: "/apps/9router/usage", Destination: "/root/.9router"},
	})
	allowed, _ := PowerAPIKeyExecCheck(info)
	assert.True(t, allowed)

	info = powerExecTestContainer(powerExecLabel(), nil, []container.MountPoint{
		{Type: mount.TypeBind, Source: "/var/run/docker.sock", Destination: "/var/run/docker.sock"},
	})
	allowed, reason := PowerAPIKeyExecCheck(info)
	assert.False(t, allowed)
	assert.NotEmpty(t, reason)
}

func Test_PowerAPIKeyExecCheck_Hardening(t *testing.T) {
	tests := []struct {
		name       string
		hostConfig *container.HostConfig
	}{
		{"privileged", &container.HostConfig{Privileged: true}},
		{"sys_admin capability", &container.HostConfig{CapAdd: []string{"sys_admin"}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			info := powerExecTestContainer(powerExecLabel(), test.hostConfig, nil)
			allowed, reason := PowerAPIKeyExecCheck(info)
			assert.False(t, allowed)
			assert.NotEmpty(t, reason)
		})
	}
}
