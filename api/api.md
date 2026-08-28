The Portainer API is an HTTP API served by Portainer. It is used by the Portainer UI, and anything you can do in the UI can also be done via the HTTP API.

API examples are available in the [Portainer documentation](https://documentation.portainer.io/api/api-examples/)

You can find out more about Portainer [on our website](http://portainer.io) and get some support on [Slack](http://portainer.io/slack/).

# Authentication

Most of the API endpoints require authentication, as well as some level of authorization.
Portainer uses JSON Web Tokens to manage authentication. You must provide a token in the **Authorization** header of each request using the **Bearer** scheme.

Example:

```
Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6MSwidXNlcm5hbWUiOiJhZG1pbiIsInJvbGUiOjEsImV4cCI6MTQ5OTM3NjE1NH0.NJ6vE8FY1WG6jsRQzfMqeatJ4vh2TWAeeYfDhP71YEE
```

# Security

Each API endpoint has an associated access policy, documented in its description.

The following policies are available:

- Public access
- Authenticated access
- Restricted access
- Administrator access

### Public access

No authentication is required.

### Authenticated access

Authentication is required.

### Restricted access

Authentication is required. Additional checks may apply to verify access to the resource, and returned data may be filtered.

### Administrator access

Authentication and an administrator role are both required.

# InfoPark fork offline discovery

This build may run in offline organizations where external documentation is not available.

Agents and local LLM tools should start with:

```
GET /llms.txt
```

For machine-readable fork capabilities, use:

```
GET /api/system/fork-capabilities
```

The capabilities document describes InfoPark fork additions, API token access presets (`disabled`, `read_only`, `power`, `manage`), Power-token allowed operations, and special endpoints such as service force update, stack webhooks, remote Portainer management, and Portainer self-update.

# Execute Docker requests

Portainer does not expose dedicated endpoints for managing Docker resources (create a container, remove a volume, etc).

Instead, it acts as a reverse-proxy to the Docker HTTP API, allowing you to execute Docker requests via the Portainer HTTP API.

To do so, use the `/endpoints/{id}/docker` endpoint. Note that this endpoint is not documented below due to Swagger limitations. It has a restricted access policy, so authentication is still required. Any request made to this endpoint is proxied to the Docker API of the associated environment - request and response objects are identical to those in the [Docker official documentation](https://docs.docker.com/engine/api).

# Fork API token access presets

This fork supports coarse API token access presets on user API keys:

- `disabled`: deny all API requests made with this token.
- `read_only`: allow read-only API requests.
- `power`: allow read-only requests plus selected operational actions such as start, stop, restart, pause, unpause, and approved service refresh operations.
- `manage`: allow full access permitted by the owning user account.

Use `PUT /api/users/{id}/tokens/{keyID}` to update a token:

```
{
  "accessPreset": "power"
}
```

The same endpoint can temporarily elevate a token. The elevation is stored on the API key and is evaluated on each API-key authentication request, so it survives a Portainer restart and automatically stops applying after the Unix timestamp expires.

```
{
  "accessPreset": "power",
  "temporaryAccessPreset": "manage",
  "temporaryAccessExpiresAt": 1710000000
}
```

Clear a temporary elevation by sending an empty `temporaryAccessPreset` and `temporaryAccessExpiresAt` set to `0`.

Token list and update responses include `effectiveAccessPreset`. Agents should use that field to know the currently enforced level instead of recomputing temporary elevation client-side.

Agents authenticated with an API key can inspect the key used for the current request:

```
GET /api/users/me/current-api-key
```

The endpoint only works with `X-API-Key` authentication and returns the same API key fields as the token list, including `effectiveAccessPreset`.

# Private Registry

When using a private registry, include a Base64-encoded JSON string in the request header. The header parameter name is `X-Registry-Auth` and the value should encode the following structure: ‘{"registryId":\<registryId\>}’ where `<registryId>` is the ID of the registry where the repository was created.

Example encoded value:

```
eyJyZWdpc3RyeUlkIjoxfQ==
```

# InfoPark API token access presets

API keys can use one of these access presets: `disabled`, `read_only`, `power`,
or `manage`.

The `power` preset allows read requests, stack start/stop, selected container
lifecycle actions, service force updates through
`PUT /api/endpoints/{id}/forceupdateservice`, and restricted Docker exec access
through Docker exec create/start/resize operations.

Restricted Docker exec access requires the target container to have this label:

```
portainer.infopark.power.exec=true
```

Even with the label, exec is denied for containers with dangerous host access,
including Docker socket mounts, privileged mode, selected host system mounts, and
`SYS_ADMIN`.
