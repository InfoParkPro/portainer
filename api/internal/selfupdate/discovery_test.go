package selfupdate

import (
	"context"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/require"
)

func TestDiscoverCurrentContainerPrefersRunningPortainerCurrent(t *testing.T) {
	client := &fakeDiscoveryClient{
		containers: []container.Summary{
			{ID: "current-id", Names: []string{"/portainer-current"}, State: "running"},
			{ID: "rollback-id", Names: []string{"/portainer-rollback-old"}, State: "running"},
		},
		inspects: map[string]container.InspectResponse{
			"current-id":   inspectResponse("current-id", "/portainer-current", true),
			"old-hostname": inspectResponse("rollback-id", "/portainer-rollback-old", true),
		},
	}

	inspect, err := DiscoverCurrentContainer(context.Background(), client, "old-hostname")

	require.NoError(t, err)
	require.Equal(t, "current-id", inspect.ID)
	require.Equal(t, []string{"list-running", "inspect current-id"}, client.calls)
}

func TestDiscoverCurrentContainerRejectsStoppedHostnameFallback(t *testing.T) {
	client := &fakeDiscoveryClient{
		inspects: map[string]container.InspectResponse{
			"old-hostname": inspectResponse("rollback-id", "/portainer-rollback-old", false),
		},
	}

	_, err := DiscoverCurrentContainer(context.Background(), client, "old-hostname")

	require.ErrorContains(t, err, "is not running")
}

func TestDiscoverCurrentContainerRejectsMultipleRunningPortainerCurrentContainers(t *testing.T) {
	client := &fakeDiscoveryClient{
		containers: []container.Summary{
			{ID: "first-id", Names: []string{"/portainer-current"}, State: "running"},
			{ID: "second-id", Names: []string{"/portainer-current"}, State: "running"},
		},
	}

	_, err := DiscoverCurrentContainer(context.Background(), client, "hostname")

	require.ErrorContains(t, err, "multiple running containers named portainer-current")
}

type fakeDiscoveryClient struct {
	containers []container.Summary
	inspects   map[string]container.InspectResponse
	calls      []string
}

func (c *fakeDiscoveryClient) ContainerList(_ context.Context, _ container.ListOptions) ([]container.Summary, error) {
	c.calls = append(c.calls, "list-running")
	return c.containers, nil
}

func (c *fakeDiscoveryClient) ContainerInspect(_ context.Context, containerID string) (container.InspectResponse, error) {
	c.calls = append(c.calls, "inspect "+containerID)
	return c.inspects[containerID], nil
}

func inspectResponse(id string, name string, running bool) container.InspectResponse {
	return container.InspectResponse{ContainerJSONBase: &container.ContainerJSONBase{
		ID:    id,
		Name:  name,
		State: &container.State{Running: running},
	}}
}
