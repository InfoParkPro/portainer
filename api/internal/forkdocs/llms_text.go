package forkdocs

import (
	"fmt"

	portainer "github.com/portainer/portainer/api"
)

// LLMSText returns the offline HTTP-only API cookbook served at /llms.txt.
// The text must stay free of backticks so it can be embedded in markdown
// pipelines without escaping.
func LLMSText() string {
	return fmt.Sprintf(`# Portainer API cookbook for autonomous agents

Scope: operate this Portainer instance (InfoPark Portainer CE fork, API version %s) over HTTP only, without source code access. The instance may run in an offline organization, so this file must be self-sufficient.

Documentation chain, in order you should consult it:
1. This file: cookbook of working requests for every common operation.
2. GET /api/system/fork-capabilities: compact machine-readable fork capabilities, API token preset rules and extra methods.
3. GET /api/docs/openapi.yaml: the COMPLETE OpenAPI (Swagger) 2.0 specification with basePath /api. It contains every route, parameter, request body and response schema of this instance. If an operation is not covered here, or a request fails in a way you do not understand, look the route up there.

## Conventions

- HOST is the instance base URL, for example https://portainer.example.com. All paths below are appended to HOST.
- Bodies and responses are JSON (Content-Type: application/json) except where multipart/form-data is stated.
- Authentication: send header X-API-Key: <token> with every request, or a JWT obtained from POST /api/auth as Authorization: Bearer <jwt>.
- Errors are JSON with a proper HTTP status, shape: {"message": "...", "details": "..."}.
- HTTP 405 means your route shape is wrong (wrong method or wrong path), never a permissions problem; permissions problems return 403.
- BASE_DOCKER = /api/endpoints/{id}/docker/{version} is the Docker Engine API proxy. {id} is an environment id from GET /api/endpoints; {version} is a Docker API version such as 1.43 (discover the exact supported value with GET BASE_DOCKER/version).
- BASE_KUBERNETES = /api/endpoints/{id}/kubernetes is the raw Kubernetes API proxy; Portainer-native Kubernetes routes live under /api/kubernetes/{id}/... (see Kubernetes section).

## Authentication and API keys

- JWT login: POST /api/auth with {"Username": "...", "Password": "..."} returns {"jwt": "..."}. POST /api/auth/logout invalidates it, POST /api/auth/refresh renews it.
- Create an API token for user {id}: POST /api/users/{id}/tokens with {"password": "<that user's password>", "description": "..."} returns the key; the field rawAPIKey is shown exactly once, save it immediately.
- API tokens carry an access preset (fork feature): disabled, read_only, power or manage. manage acts as the token owner; power is read-only plus a safe-operation allowlist; read_only allows GET/HEAD/OPTIONS only; disabled rejects everything.
- Inspect the preset of the token you are using: GET /api/users/me/current-api-key returns accessPreset and effectiveAccessPreset.
- Change a preset (admin or token owner): PUT /api/users/{id}/tokens/{keyID} with {"accessPreset": "power"}.
- A denied operation returns 403, never 405.

## Discovery checklist

Run these first when you know nothing about the instance:
- GET /api/status - instance version and database version.
- GET /api/system/version - build details.
- GET /api/system/fork-capabilities - fork capabilities, presets, extra methods.
- GET /api/docs/openapi.yaml - full API specification.
- GET /api/users/me - who am I; Role 1 is admin, 2 is regular user.
- GET /api/endpoints - environments to operate on.
- GET /api/settings/public - unauthenticated settings summary.

## Environments

- List: GET /api/endpoints. Each item has Id, Name and Type. Types: 1 standalone Docker, 2 Docker agent, 3 Azure, 4 Edge agent on Docker, 5 local Kubernetes, 6 Kubernetes agent, 7 Edge agent on Kubernetes.
- Inspect: GET /api/endpoints/{id}. Update: PUT /api/endpoints/{id}. Delete: DELETE /api/endpoints/{id}.
- Create (multipart/form-data, not JSON): POST /api/endpoints with fields Name and EndpointCreationType (1 local Docker, 2 agent, 3 Azure, 4 Edge agent, 5 local Kubernetes), plus URL, PublicURL, GroupID, ContainerEngine (docker or podman), TLS, TLSSkipVerify, TLSSkipClientVerify and optional TLS certificate file fields.
- Snapshots: POST /api/endpoints/snapshot refreshes stored snapshots for all environments.
- Groups: GET /api/endpoint_groups, POST /api/endpoint_groups ({"Name": "...", "Endpoints": [ids]}); manage: GET/PUT/DELETE /api/endpoint_groups/{id}; membership: PUT/DELETE /api/endpoint_groups/{id}/endpoints/{endpointId}.
- Tags: GET /api/tags, POST /api/tags ({"Name": "..."}), DELETE /api/tags/{id}.

## Docker operations (Docker Engine API proxy)

Requests through BASE_DOCKER behave like the Docker Engine API of the target host for version {version}; bodies and query parameters follow that API.

Containers:
- List: GET BASE_DOCKER/containers/json?all=true
- Inspect: GET BASE_DOCKER/containers/{containerID}/json
- Logs: GET BASE_DOCKER/containers/{containerID}/logs?stdout=true&stderr=true&tail=200
- Lifecycle: POST .../start, .../stop, .../restart, .../kill, .../pause, .../unpause
- Rename: POST BASE_DOCKER/containers/{containerID}/rename?name=newname
- Delete: DELETE BASE_DOCKER/containers/{containerID}?force=true
- Run a command (two calls):
  1. POST BASE_DOCKER/containers/{containerID}/exec with {"AttachStdout": true, "AttachStderr": true, "Cmd": ["sh", "-c", "uptime"]} returns {"Id": "<execID>"}
  2. POST BASE_DOCKER/exec/{execID}/start with {"Detach": false, "Tty": false}; the raw output stream is the response body.
  Power preset tokens may exec only into containers labeled portainer.infopark.power.exec=true that pass safety checks; privileged containers, containers exposing docker.sock, with dangerous host mounts or SYS_ADMIN capability are always refused.

Images:
- List: GET BASE_DOCKER/images/json
- Pull: POST BASE_DOCKER/images/create?fromImage=nginx&tag=alpine
- Delete: DELETE BASE_DOCKER/images/{imageID}?force=true

Volumes:
- List: GET BASE_DOCKER/volumes; Create: POST BASE_DOCKER/volumes ({"Name": "...", "Driver": "local"}); Delete: DELETE BASE_DOCKER/volumes/{name}

Networks:
- List: GET BASE_DOCKER/networks; Create: POST BASE_DOCKER/networks; Inspect: GET BASE_DOCKER/networks/{networkID}; Delete: DELETE BASE_DOCKER/networks/{networkID}

Swarm (when the environment runs Swarm):
- Cluster id and state: GET BASE_DOCKER/swarm (use the ID field as SwarmID for stack creation)
- Nodes: GET BASE_DOCKER/nodes
- Services: GET BASE_DOCKER/services. Generic update POST BASE_DOCKER/services/{serviceID}/update is refused for power tokens; on this fork use PUT /api/endpoints/{id}/forceupdateservice with {"serviceID": "<id or name>", "pullImage": true}.

## Stacks (Docker Compose, Swarm and Kubernetes)

IMPORTANT route shape: stack creation exists ONLY at POST /api/stacks/create/{type}/{method}. POST /api/stacks is the list route and answers 405 to POST - that is by design, not a fork bug. Types: standalone (compose on a single host), swarm, kubernetes. Methods: string (inline content), file (multipart upload), repository (git), url.

- List: GET /api/stacks. Inspect: GET /api/stacks/{id}. Delete: DELETE /api/stacks/{id}.
- Delete a Kubernetes stack by name: DELETE /api/stacks/name/{name}.
- Read the current compose content: GET /api/stacks/{id}/file.
- Start/stop: POST /api/stacks/{id}/start and POST /api/stacks/{id}/stop.
- Create a standalone stack from inline content: POST /api/stacks/create/standalone/string?endpointId={id} with {"Name": "myapp", "StackFileContent": "services:\n  web:\n    image: nginx:alpine", "Env": [{"name": "KEY", "value": "val"}]}.
- Create a Swarm stack from inline content: POST /api/stacks/create/swarm/string?endpointId={id} - same body plus "SwarmID" taken from GET BASE_DOCKER/swarm (ID field).
- Create Kubernetes stacks: POST /api/stacks/create/kubernetes/string?endpointId={id} with StackFileContent containing the manifest; exact payload fields in /api/docs/openapi.yaml.
- Create from a git repository: POST /api/stacks/create/standalone/repository?endpointId={id} with {"Name": "...", "RepositoryURL": "https://...", "RepositoryReferenceName": "refs/heads/main", "ComposeFile": "docker-compose.yml", "Env": [...], "AutoUpdate": {...}}. The same route exists for swarm. RepositoryAuthentication with Username and Password is deprecated; see git credential fields in /api/docs/openapi.yaml.
- Update an inline stack: PUT /api/stacks/{id} with {"StackFileContent": "..."} plus optional Env, Prune, RepullImageAndRedeploy, Webhook.
- Redeploy a git-based stack: PUT /api/stacks/{id}/git/redeploy with {"RepositoryReferenceName": "...", "Prune": true, "RepullImageAndRedeploy": true} (optional RepositoryAuthentication).
- Switch a file-based stack to git: POST /api/stacks/{id}/git with {"RepositoryURL": "...", "RepositoryReferenceName": "...", "ComposeFile": "..."}.
- Migrate a stack to another endpoint: POST /api/stacks/{id}/migrate; associate an external stack: PUT /api/stacks/{id}/associate.
- Stack redeploy webhook: enable with PUT /api/stacks/{id} and "Webhook": true, then trigger (public, no auth) POST /api/stacks/webhooks/{webhookID}; throttled to one run per 10 minutes.
- Regular (non-admin) users may manage stacks only when the endpoint SecuritySettings.AllowStackManagementForRegularUsers is enabled; otherwise the API returns 403.

## Edge compute

- Edge stacks: POST /api/edge_stacks/create/string with {"Name": "...", "StackFileContent": "...", "DeploymentType": 0, "EdgeGroups": [1]}. DeploymentType 0 is compose, 1 is a Kubernetes manifest. EdgeGroups are ids from GET /api/edge_groups. Methods file and repository exist too: POST /api/edge_stacks/create/file and POST /api/edge_stacks/create/repository (git; FilePathInRepository defaults to docker-compose.yml or the kubernetes manifest).
- Manage: GET /api/edge_stacks, GET/PUT/DELETE /api/edge_stacks/{id}, file content: GET /api/edge_stacks/{id}/file.
- Edge groups: GET /api/edge_groups, POST /api/edge_groups ({"Name": "...", "Dynamic": false, "Endpoints": [ids]}).
- Edge jobs (scripts executed on edge devices): create via POST /api/edge_jobs/create/string or /api/edge_jobs/create/file; exact fields in /api/docs/openapi.yaml.

## Kubernetes

- Portainer-native routes: /api/kubernetes/{id}/<resource> where {id} is the environment id. Resources: applications, applications/count, configmaps, configmaps/count, cron_jobs (delete via POST /cron_jobs/delete), jobs, events, ingresses (+/count), ingresscontrollers, namespaces (GET list, POST create, PUT/DELETE /namespaces/{namespace}), namespaces/{namespace}/system (PUT toggles system state, admin), services (+/delete), secrets, volumes, persistent_volumes, persistent_volume_claims (+/delete, +/resize), storage_classes (+/{name}, +/{name}/default), service_accounts, roles, role_bindings, cluster_roles, cluster_role_bindings (deletes via POST /<resource>/delete), dashboard, rbac_enabled, version, nodes/{name}/drain, describe, metrics/nodes, metrics/pods/namespace/{namespace}. Exact query parameters and bodies: /api/docs/openapi.yaml.
- Kubeconfig for the current user: GET /api/kubernetes/config.
- Raw cluster access: anything under BASE_KUBERNETES (for example BASE_KUBERNETES/api/v1/namespaces) is proxied to the cluster API server.
- Helm: list releases GET /api/endpoints/{id}/kubernetes/helm; install POST /api/endpoints/{id}/kubernetes/helm; inspect GET /api/endpoints/{id}/kubernetes/helm/{release}; history GET /api/endpoints/{id}/kubernetes/helm/{release}/history; rollback POST /api/endpoints/{id}/kubernetes/helm/{release}/rollback; uninstall DELETE /api/endpoints/{id}/kubernetes/helm/{release}. Chart search: GET /api/templates/helm; show chart/values/readme: GET /api/templates/helm/{command}.

## Registries

- Create/list: POST/GET /api/registries ({"Name": "...", "URL": "...", "Authentication": true, "Username": "...", "Password": "..."}). Management routes are admin-only.
- Manage: GET/PUT/DELETE /api/registries/{id}; POST /api/registries/{id}/configure; connectivity check POST /api/registries/ping.
- Inspect (needs registry access): GET /api/registries/{id}; GitLab proxy: GET /api/registries/proxies/gitlab/...

## Users, teams and access

- Create user: POST /api/users {"Username": "...", "Password": "...", "Role": 2} (Role 1 admin, 2 regular). List: GET /api/users; update: PUT /api/users/{id}; delete: DELETE /api/users/{id}; admin check/init: GET /api/users/admin/check, POST /api/users/admin/init.
- Teams: GET /api/teams, POST /api/teams ({"Name": "...", "TeamLeaders": [userIds]}); manage: GET/PUT/DELETE /api/teams/{id}; members: GET /api/teams/{id}/memberships.
- Memberships: GET /api/team_memberships, POST /api/team_memberships ({"TeamID": 1, "UserID": 2, "Role": 2}) (Role 1 team leader, 2 regular); DELETE /api/team_memberships/{id}.
- API tokens: see Authentication and API keys.
- Custom RBAC roles: GET /api/roles.
- LDAP check: POST /api/ldap/check.

## Resource controls (per-resource access)

Restrict who may see or use a specific resource:
- POST /api/resource_controls {"ResourceID": "<docker resource id or stack name>", "Type": 6, "Users": [], "Teams": [1], "Public": false, "AdministratorsOnly": false}
- Types: 1 container, 2 service, 3 volume, 4 network, 5 secret, 6 stack, 7 config, 8 custom template, 9 azure.
- Update: PUT /api/resource_controls/{id}; delete: DELETE /api/resource_controls/{id}.

## Custom templates

- Create from inline content: POST /api/custom_templates/create/string {"Title": "...", "FileContent": "services:...", "Type": 2, "Platform": 1, "Description": "...", "Note": "...", "Logo": "https://..."}. Type: 1 swarm, 2 compose, 3 kubernetes. Platform: 1 linux, 2 windows.
- Create from file (multipart) or git: POST /api/custom_templates/create/file and POST /api/custom_templates/create/repository.
- Manage: GET /api/custom_templates, GET/PUT/DELETE /api/custom_templates/{id}.

## Webhooks (container or service redeploy)

- Create: POST /api/webhooks {"EndpointID": 1, "ResourceID": "<container or service id>", "WebhookType": 2} (WebhookType 1 service, 2 container). The response carries the public token.
- Invoke (public, no auth): POST /api/webhooks/{token} - pulls the image and redeploys the resource; throttled to one run per 10 minutes.
- List: GET /api/webhooks; update registry: PUT /api/webhooks/{id}; delete: DELETE /api/webhooks/{id}.

## Settings, backup and system

- Settings: GET /api/settings, PUT /api/settings; public subset without auth: GET /api/settings/public.
- TLS certificate configuration: GET/PUT /api/ssl (admin).
- Backup: POST /api/backup {"Password": "..."} returns a backup archive (admin). Restore: POST /api/restore (multipart, public during initial setup, see spec).
- System: GET /api/system/info, GET /api/system/nodes, GET /api/system/version; message of the day: GET /api/motd.
- Upload helper for TLS files (multipart, admin): POST /api/upload/tls/ca, /api/upload/tls/cert, /api/upload/tls/key.
- App templates: GET /api/templates; template file: POST /api/templates/{id}/file.

## Fork extras

- Remote Portainer instances: GET/POST /api/remote_portainers, PUT/DELETE /api/remote_portainers/{id} (admin; stored tokens are never returned). Fields: name, url, apiToken, tlsskipverify.
- Self-update: GET /api/system/self-update/plan (admin) checks whether this instance runs as a plain Docker container; POST /api/system/self-update/start {"targetImage": "..."} starts a helper container that replaces it and keeps the old container for manual rollback.
- Current API key introspection: GET /api/users/me/current-api-key.
- WebSocket routes (/api/websocket/attach, /api/websocket/exec, /api/websocket/pod, /api/websocket/kubernetes-shell) serve browser sessions; agents should use the Docker proxy exec flow instead.

## Common gotchas

- 405 on POST /api/stacks is correct: creation is POST /api/stacks/create/{type}/{method}.
- 403 means: token preset denies the operation, admin-only route with a regular user, stack management disabled for regular users on that endpoint, or missing access to the resource.
- rawAPIKey from POST /api/users/{id}/tokens is displayed once only.
- POST /api/endpoints and every /create/file method consume multipart/form-data, not JSON.
- Power tokens cannot DELETE containers, stacks, services, volumes, networks, images, secrets or configs; cannot update services generically (use PUT /api/endpoints/{id}/forceupdateservice); exec only into containers with label portainer.infopark.power.exec=true.
- Swagger-2.0-style proxy note: Docker proxy responses match the Docker Engine API, so unknown fields are usually platform fields documented by Docker, not Portainer errors.

## Full reference

GET /api/docs/openapi.yaml returns the complete OpenAPI 2.0 specification (basePath /api) with every route, parameter and schema under definitions. Search it for any path, query parameter or response field not detailed above. GET /api/system/fork-capabilities returns the compact JSON version of this guide plus fork-specific capability notes.
`, portainer.APIVersion)
}
