package forkdocs

import portainer "github.com/portainer/portainer/api"

type CapabilityDocument struct {
	Fork          string                  `json:"fork"`
	Version       string                  `json:"version"`
	Description   string                  `json:"description"`
	Discovery     []string                `json:"discovery"`
	AccessPresets map[string]AccessPreset `json:"accessPresets"`
	Methods       []MethodCapability      `json:"methods"`
	DockerProxy   []DockerProxyCapability `json:"dockerProxy"`
	Notes         []string                `json:"notes"`
}

type AccessPreset struct {
	Description string   `json:"description"`
	Allowed     []string `json:"allowed"`
	Denied      []string `json:"denied,omitempty"`
}

type MethodCapability struct {
	Method      string            `json:"method"`
	Path        string            `json:"path"`
	Access      string            `json:"access"`
	Description string            `json:"description"`
	Body        map[string]string `json:"body,omitempty"`
	Example     string            `json:"example,omitempty"`
}

type DockerProxyCapability struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Access      string `json:"access"`
	Description string `json:"description"`
}

func Capabilities() CapabilityDocument {
	return CapabilityDocument{
		Fork:        "infopark-portainer",
		Version:     portainer.APIVersion,
		Description: "Offline machine-readable capabilities for the InfoPark Portainer CE fork.",
		Discovery: []string{
			"GET /llms.txt",
			"GET /api/system/fork-capabilities",
			"GET /api/docs/openapi.yaml",
			"GET /api/users/me/current-api-key",
			"GET /api/system/version",
			"GET /api/status",
		},
		AccessPresets: map[string]AccessPreset{
			"disabled": {
				Description: "API token access is disabled.",
				Allowed:     []string{},
			},
			"read_only": {
				Description: "Read-only API token access.",
				Allowed: []string{
					"GET non-websocket API routes",
					"HEAD non-websocket API routes",
					"OPTIONS non-websocket API routes",
				},
				Denied: []string{
					"GET /api/websocket/*",
					"POST, PUT, PATCH, DELETE API routes",
				},
			},
			"power": {
				Description: "Read-only plus safe operational actions.",
				Allowed: []string{
					"GET non-websocket API routes",
					"HEAD non-websocket API routes",
					"OPTIONS non-websocket API routes",
					"POST /api/stacks/{id}/start",
					"POST /api/stacks/{id}/stop",
					"POST /api/endpoints/{id}/docker/{version}/containers/{containerID}/start",
					"POST /api/endpoints/{id}/docker/{version}/containers/{containerID}/stop",
					"POST /api/endpoints/{id}/docker/{version}/containers/{containerID}/restart",
					"POST /api/endpoints/{id}/docker/{version}/containers/{containerID}/kill",
					"POST /api/endpoints/{id}/docker/{version}/containers/{containerID}/pause",
					"POST /api/endpoints/{id}/docker/{version}/containers/{containerID}/unpause",
					"POST /api/endpoints/{id}/docker/{version}/containers/{containerID}/exec when the container has label portainer.infopark.power.exec=true and passes safety checks",
					"POST /api/endpoints/{id}/docker/{version}/exec/{execID}/start when the exec target container passes Power exec safety checks",
					"POST /api/endpoints/{id}/docker/{version}/exec/{execID}/resize when the exec target container passes Power exec safety checks",
					"GET /api/websocket/exec for exec sessions created through an allowed container exec request",
					"PUT /api/endpoints/{id}/forceupdateservice",
				},
				Denied: []string{
					"GET /api/websocket/* except /api/websocket/exec",
					"Docker exec for containers without label portainer.infopark.power.exec=true",
					"Docker exec for privileged containers, containers with docker.sock or dangerous host mounts, or containers with SYS_ADMIN capability",
					"DELETE containers, stacks, services, volumes, networks, images, secrets, configs",
					"POST /api/endpoints/{id}/docker/{version}/services/{serviceID}/update",
					"POST /api/stacks",
					"PUT /api/stacks/{id}",
				},
			},
			"manage": {
				Description: "Full API token access equivalent to the token owner permissions.",
				Allowed:     []string{"All API routes allowed by the token owner permissions."},
			},
		},
		Methods: []MethodCapability{
			{
				Method:      "GET",
				Path:        "/api/users/me/current-api-key",
				Access:      "API key authentication",
				Description: "Inspect the API key used for the current request. Returns accessPreset, temporaryAccessPreset, temporaryAccessExpiresAt, and effectiveAccessPreset.",
				Example:     `curl "https://portainer.example.com/api/users/me/current-api-key" -H "X-API-Key: TOKEN"`,
			},
			{
				Method:      "GET",
				Path:        "/api/endpoints/{id}/docker/{version}/swarm",
				Access:      "Power, Manage",
				Description: "Inspect the Swarm cluster of an environment. The ID field of the response is the SwarmID required to create Swarm stacks.",
				Example:     `curl "https://portainer.example.com/api/endpoints/1/docker/1.43/swarm" -H "X-API-Key: TOKEN"`,
			},
			{
				Method:      "POST",
				Path:        "/api/stacks/create/standalone/string",
				Access:      "Manage, Admin",
				Description: "Create a standalone compose stack from inline content. Requires the endpointId query parameter. Stack creation only exists under /api/stacks/create/{type}/{method}; POST /api/stacks is the list route and answers 405.",
				Body: map[string]string{
					"Name":             "Stack name.",
					"StackFileContent": "Compose file content.",
					"Env":              "Optional array of {name, value}.",
				},
				Example: `curl -X POST "https://portainer.example.com/api/stacks/create/standalone/string?endpointId=1" -H "X-API-Key: TOKEN" -H "Content-Type: application/json" --data '{"Name":"demo","StackFileContent":"services:\n  web:\n    image: nginx:alpine","Env":[]}'`,
			},
			{
				Method:      "POST",
				Path:        "/api/stacks/create/swarm/string",
				Access:      "Manage, Admin",
				Description: "Create a Swarm stack from inline content. Requires the endpointId query parameter and SwarmID in the body from GET /api/endpoints/{id}/docker/{version}/swarm.",
				Body: map[string]string{
					"Name":             "Stack name.",
					"StackFileContent": "Compose file content.",
					"SwarmID":          "Swarm cluster ID from GET /api/endpoints/{id}/docker/{version}/swarm.",
					"Env":              "Optional array of {name, value}.",
				},
				Example: `curl -X POST "https://portainer.example.com/api/stacks/create/swarm/string?endpointId=1" -H "X-API-Key: TOKEN" -H "Content-Type: application/json" --data '{"Name":"demo","StackFileContent":"services:\n  web:\n    image: nginx:alpine","SwarmID":"swarm-id-from-swarm-inspect","Env":[]}'`,
			},
			{
				Method:      "PUT",
				Path:        "/api/stacks/{id}",
				Access:      "Manage, Admin",
				Description: "Update a standalone or swarm stack. StackFileContent is required; optional Env, Prune, RepullImageAndRedeploy and Webhook.",
				Body: map[string]string{
					"StackFileContent":       "New compose file content.",
					"Env":                    "Optional array of {name, value}.",
					"Prune":                  "Boolean. Prune services that are no longer in the compose file (swarm).",
					"RepullImageAndRedeploy": "Boolean. Force image pull and redeploy.",
					"Webhook":                "Boolean. Enable the public redeploy webhook.",
				},
			},
			{
				Method:      "PUT",
				Path:        "/api/endpoints/{id}/forceupdateservice",
				Access:      "Power, Manage",
				Description: "Force update a Docker Swarm service. Set pullImage to true to pull the same tag again, for example latest.",
				Body: map[string]string{
					"serviceID": "Docker service ID or name.",
					"pullImage": "Boolean. If true, remove the image digest and query the registry during service update.",
				},
				Example: `curl -X PUT "https://portainer.example.com/api/endpoints/1/forceupdateservice" -H "X-API-Key: TOKEN" -H "Content-Type: application/json" --data '{"serviceID":"app_web","pullImage":true}'`,
			},
			{
				Method:      "POST",
				Path:        "/api/stacks/webhooks/{webhookID}",
				Access:      "Public webhook secret URL",
				Description: "Redeploy a file-based stack with image pull. Accepted calls are throttled to one run per 10 minutes per stack webhook.",
			},
			{
				Method:      "GET",
				Path:        "/api/remote_portainers",
				Access:      "Admin",
				Description: "List configured remote Portainer instances. API tokens are never returned.",
			},
			{
				Method:      "POST",
				Path:        "/api/remote_portainers",
				Access:      "Admin",
				Description: "Create a remote Portainer instance using a manually provided API token.",
				Body: map[string]string{
					"name":          "Display name.",
					"url":           "Remote Portainer base URL.",
					"apiToken":      "Remote Portainer API token.",
					"tlsskipverify": "Boolean. Skip TLS certificate verification for the remote Portainer.",
				},
			},
			{
				Method:      "PUT",
				Path:        "/api/remote_portainers/{id}",
				Access:      "Admin",
				Description: "Update a remote Portainer instance. If apiToken is omitted, the stored token is kept.",
			},
			{
				Method:      "DELETE",
				Path:        "/api/remote_portainers/{id}",
				Access:      "Admin",
				Description: "Delete a remote Portainer connection.",
			},
			{
				Method:      "GET",
				Path:        "/api/system/self-update/plan",
				Access:      "Admin",
				Description: "Inspect whether this Portainer instance can self-update as a plain Docker container.",
			},
			{
				Method:      "POST",
				Path:        "/api/system/self-update/start",
				Access:      "Admin",
				Description: "Start a helper container that replaces a plain Docker Portainer container and leaves the old container for manual rollback.",
				Body: map[string]string{
					"targetImage": "Replacement Portainer image. Empty means use the current image.",
				},
			},
			{
				Method:      "PUT",
				Path:        "/api/users/{id}/tokens/{keyID}",
				Access:      "Restricted",
				Description: "Update an API token access preset.",
				Body: map[string]string{
					"accessPreset": "One of disabled, read_only, power, manage.",
				},
			},
			{
				Method:      "POST",
				Path:        "/api/endpoints/{id}/docker/{version}/containers/{containerID}/exec",
				Access:      "Power, Manage",
				Description: "Create a Docker exec session. Power tokens require container label portainer.infopark.power.exec=true and runtime safety checks.",
			},
		},
		DockerProxy: []DockerProxyCapability{
			{
				Method:      "POST",
				Path:        "/api/endpoints/{id}/docker/{version}/containers/{containerID}/restart",
				Access:      "Power, Manage",
				Description: "Restart a Docker container through Portainer's Docker proxy.",
			},
			{
				Method:      "POST",
				Path:        "/api/endpoints/{id}/docker/{version}/containers/{containerID}/start",
				Access:      "Power, Manage",
				Description: "Start a Docker container through Portainer's Docker proxy.",
			},
			{
				Method:      "POST",
				Path:        "/api/endpoints/{id}/docker/{version}/containers/{containerID}/stop",
				Access:      "Power, Manage",
				Description: "Stop a Docker container through Portainer's Docker proxy.",
			},
			{
				Method:      "POST",
				Path:        "/api/endpoints/{id}/docker/{version}/containers/{containerID}/kill",
				Access:      "Power, Manage",
				Description: "Kill a Docker container through Portainer's Docker proxy.",
			},
			{
				Method:      "POST",
				Path:        "/api/endpoints/{id}/docker/{version}/containers/{containerID}/pause",
				Access:      "Power, Manage",
				Description: "Pause a Docker container through Portainer's Docker proxy.",
			},
			{
				Method:      "POST",
				Path:        "/api/endpoints/{id}/docker/{version}/containers/{containerID}/unpause",
				Access:      "Power, Manage",
				Description: "Unpause a Docker container through Portainer's Docker proxy.",
			},
			{
				Method:      "POST",
				Path:        "/api/endpoints/{id}/docker/{version}/containers/{containerID}/exec",
				Access:      "Power, Manage",
				Description: "Create an exec session. Power requires label portainer.infopark.power.exec=true and denies privileged containers, docker.sock, dangerous host mounts, and SYS_ADMIN.",
			},
			{
				Method:      "POST",
				Path:        "/api/endpoints/{id}/docker/{version}/exec/{execID}/start",
				Access:      "Power, Manage",
				Description: "Start an exec session. Power re-checks the exec target container before starting.",
			},
			{
				Method:      "POST",
				Path:        "/api/endpoints/{id}/docker/{version}/exec/{execID}/resize",
				Access:      "Power, Manage",
				Description: "Resize an exec session. Power re-checks the exec target container.",
			},
		},
		Notes: []string{
			"This document is intentionally compact for offline agents and local LLMs.",
			"Stack creation is only available at POST /api/stacks/create/{type}/{method}; POST /api/stacks is the list route and returns 405. The full agent cookbook with every route is at GET /llms.txt and the complete OpenAPI 2.0 specification at GET /api/docs/openapi.yaml.",
			"Do not call generic Docker service update with Power tokens; use /api/endpoints/{id}/forceupdateservice instead.",
			"Power token Docker exec checks are performed at request time against the current container inspect data, not from a cached allowlist.",
			"Docker proxy request and response bodies follow the Docker Engine API for the selected API version.",
			"External documentation links may be unreachable in offline organizations.",
		},
	}
}
