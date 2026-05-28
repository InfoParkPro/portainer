# Remote Portainer API Token Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add a separate Remote Portainer feature that lets an admin connect to another Portainer instance with a manually created API token, list its stacks, view stack files, and update remote stacks while all stack data remains on the remote Portainer.

**Architecture:** Do not model remote Portainer instances as environments. Store only remote connection records locally, then use a dedicated backend client to call the remote Portainer HTTP API. Add a separate React administration area for remote Portainers so the existing Docker environment, agent, and local stack flows stay isolated.

**Tech Stack:** Go HTTP handlers with Gorilla mux, existing Portainer datastore/base services, Go `net/http` remote client, Angular state registration for React views, React Query, existing Portainer UI components.

---

## Scope

Build the MVP in one overlay branch, `local/remote-portainer-api-token`.

Included:
- Local CRUD for remote Portainer connections.
- Manual API token storage.
- Test connection against a remote Portainer.
- Remote stack list.
- Remote stack file read.
- Remote stack update for file-based stacks through the remote Portainer API.

Excluded from MVP:
- Login/password flow.
- Importing remote stacks into the local `Stack` table.
- Adding a new `EndpointType`.
- Full remote Docker runtime proxy for containers, services, logs, exec, volumes, networks.
- Automatic updates, bulk updates, and primary/standby diff.

## Overlay Rules

- Keep `main` or upstream mirror clean.
- Put this feature in one autonomous branch: `local/remote-portainer-api-token`.
- Do not split backend and UI into dependent branches.
- Prefer new files and packages.
- Touch existing files only for service registration, route registration, and sidebar/state registration.
- Do not change the existing `/api/stacks` semantics.
- Do not make remote stacks look like local stack records.

## Proposed API

Local Portainer API:

```text
GET    /api/remote_portainers
POST   /api/remote_portainers
GET    /api/remote_portainers/{id}
PUT    /api/remote_portainers/{id}
DELETE /api/remote_portainers/{id}

POST   /api/remote_portainers/{id}/test

GET    /api/remote_portainers/{id}/stacks
GET    /api/remote_portainers/{id}/stacks/{stackId}
GET    /api/remote_portainers/{id}/stacks/{stackId}/file
PUT    /api/remote_portainers/{id}/stacks/{stackId}
```

Remote calls made by the backend client:

```text
GET /api/status
GET /api/stacks
GET /api/stacks/{stackId}
GET /api/stacks/{stackId}/file
PUT /api/stacks/{stackId}?endpointId=<remote stack EndpointId>
```

Use `X-API-KEY: <token>` for remote authentication.

## Data Model

Add to `api/portainer.go`:

```go
type RemotePortainerID int

type RemotePortainer struct {
    ID            RemotePortainerID `json:"Id"`
    Name          string            `json:"Name"`
    URL           string            `json:"URL"`
    APIToken      string            `json:"-"`
    TLSSkipVerify bool              `json:"TLSSkipVerify"`
    CreatedAt     int64             `json:"CreatedAt"`
    UpdatedAt     int64             `json:"UpdatedAt"`
}
```

Note: hide `APIToken` from list/read responses. Accept it on create/update payloads. If an existing Portainer encryption helper is available and fits this use case, encrypt before persistence. Otherwise keep this as a known MVP security debt and document it in the branch notes.

## Task 1: Datastore Service

**Files:**
- Modify: `api/portainer.go`
- Create: `api/dataservices/remoteportainer/remoteportainer.go`
- Create: `api/dataservices/remoteportainer/tx.go`
- Modify: `api/dataservices/interface.go`
- Modify: `api/datastore/services.go`
- Modify: `api/datastore/services_tx.go`

**Step 1: Add model and service interface**

Add `RemotePortainerID` and `RemotePortainer` to `api/portainer.go`.

Add to `api/dataservices/interface.go`:

```go
RemotePortainer() RemotePortainerService
```

Add:

```go
type RemotePortainerService interface {
    BaseCRUD[portainer.RemotePortainer, portainer.RemotePortainerID]
    GetNextIdentifier() int
}
```

**Step 2: Implement datastore service**

Create bucket service:

```go
const BucketName = "remote_portainers"

type Service struct {
    dataservices.BaseDataService[portainer.RemotePortainer, portainer.RemotePortainerID]
}
```

Follow the shape of `api/dataservices/tag/tag.go`.

**Step 3: Register service**

In `api/datastore/services.go`, add:
- import `remoteportainer`;
- `RemotePortainerService *remoteportainer.Service`;
- initialization in `initServices`;
- `RemotePortainer()` getter.

In `api/datastore/services_tx.go`, add the transaction getter.

**Step 4: Test**

Add focused service tests if the local pattern is lightweight. Otherwise rely on handler tests in Task 3.

Run:

```bash
go test ./api/dataservices/remoteportainer ./api/datastore
```

Expected: PASS.

**Step 5: Commit**

```bash
git add api/portainer.go api/dataservices/interface.go api/dataservices/remoteportainer api/datastore/services.go api/datastore/services_tx.go
git commit -m "local: add remote portainer datastore service"
```

## Task 2: Remote Portainer HTTP Client

**Files:**
- Create: `api/remoteportainer/client/client.go`
- Create: `api/remoteportainer/client/client_test.go`

**Step 1: Define client**

Implement a small client with:
- base URL normalization;
- API token header;
- timeout;
- TLS skip verify option;
- JSON request/response handling;
- error propagation with remote status code and response body summary.

Core methods:

```go
type Client struct {
    baseURL    string
    apiToken   string
    httpClient *http.Client
}

func New(baseURL string, apiToken string, tlsSkipVerify bool) (*Client, error)
func (c *Client) Status(ctx context.Context) (*StatusResponse, error)
func (c *Client) Stacks(ctx context.Context) ([]portainer.Stack, error)
func (c *Client) Stack(ctx context.Context, id portainer.StackID) (*portainer.Stack, error)
func (c *Client) StackFile(ctx context.Context, id portainer.StackID) (string, error)
func (c *Client) UpdateStack(ctx context.Context, stack *portainer.Stack, payload UpdateStackPayload) (*portainer.Stack, error)
```

**Step 2: Add tests with `httptest.Server`**

Test:
- `X-API-KEY` is sent.
- `/api/status` is called by `Status`.
- `UpdateStack` calls `/api/stacks/{id}?endpointId=<remote EndpointId>`.
- Non-2xx responses return a useful error.

Run:

```bash
go test ./api/remoteportainer/client
```

Expected: PASS.

**Step 3: Commit**

```bash
git add api/remoteportainer/client
git commit -m "local: add remote portainer API client"
```

## Task 3: Backend Handler

**Files:**
- Create: `api/http/handler/remoteportainers/handler.go`
- Create: `api/http/handler/remoteportainers/payloads.go`
- Create: `api/http/handler/remoteportainers/remote_portainer_list.go`
- Create: `api/http/handler/remoteportainers/remote_portainer_create.go`
- Create: `api/http/handler/remoteportainers/remote_portainer_update.go`
- Create: `api/http/handler/remoteportainers/remote_portainer_delete.go`
- Create: `api/http/handler/remoteportainers/remote_portainer_test_connection.go`
- Create: `api/http/handler/remoteportainers/remote_stack_list.go`
- Create: `api/http/handler/remoteportainers/remote_stack_file.go`
- Create: `api/http/handler/remoteportainers/remote_stack_update.go`
- Create: `api/http/handler/remoteportainers/handler_test.go`
- Modify: `api/http/handler/handler.go`
- Modify: `api/http/server.go`

**Step 1: Register handler**

Add `RemotePortainerHandler *remoteportainers.Handler` to `api/http/handler/handler.go`.

Route before the catch-all file handler:

```go
case strings.HasPrefix(r.URL.Path, "/api/remote_portainers"):
    http.StripPrefix("/api", h.RemotePortainerHandler).ServeHTTP(w, r)
```

In `api/http/server.go`, instantiate:

```go
var remotePortainerHandler = remoteportainers.NewHandler(requestBouncer, server.DataStore)
```

and add it to `handler.Handler`.

**Step 2: Implement local CRUD**

Use `bouncer.AdminAccess` for all MVP routes. This keeps the first version simple and avoids inventing remote RBAC mapping.

Payload:

```go
type createRemotePortainerPayload struct {
    Name          string
    URL           string
    APIToken      string
    TLSSkipVerify bool
}
```

Validation:
- `Name` required.
- `URL` required and parseable.
- `APIToken` required on create.
- `APIToken` optional on update; empty means keep existing token.

Do not return `APIToken` in responses.

**Step 3: Implement test connection**

`POST /remote_portainers/{id}/test` loads the connection and calls remote `/api/status`.

Return:

```json
{
  "Status": "ok",
  "Version": "2.42.0"
}
```

or a 502-style handler error when the remote Portainer is unreachable or rejects the token.

**Step 4: Implement remote stack endpoints**

`GET /remote_portainers/{id}/stacks` calls remote `/api/stacks`.

`GET /remote_portainers/{id}/stacks/{stackId}` calls remote `/api/stacks/{stackId}`.

`GET /remote_portainers/{id}/stacks/{stackId}/file` calls remote `/api/stacks/{stackId}/file`.

`PUT /remote_portainers/{id}/stacks/{stackId}`:
- fetch remote stack by id;
- require `remoteStack.EndpointID != 0`;
- send update to remote `/api/stacks/{stackId}?endpointId=<remoteStack.EndpointID>`;
- pass `StackFileContent`, `Env`, `Prune`, `RepullImageAndRedeploy`.

**Step 5: Handler tests**

Use `httptest.Server` for the fake remote Portainer and a test datastore.

Test:
- create hides token in response;
- update without token keeps prior token;
- test connection uses configured token;
- remote stack update uses remote stack `EndpointId`, not local data;
- failed remote response maps to an error.

Run:

```bash
go test ./api/http/handler/remoteportainers
go test ./api/http/handler
```

Expected: PASS.

**Step 6: Commit**

```bash
git add api/http/handler/remoteportainers api/http/handler/handler.go api/http/server.go
git commit -m "local: add remote portainer API handlers"
```

## Task 4: Frontend API Layer

**Files:**
- Create: `app/react/portainer/remote-portainers/types.ts`
- Create: `app/react/portainer/remote-portainers/queries/queryKeys.ts`
- Create: `app/react/portainer/remote-portainers/queries/useRemotePortainers.ts`
- Create: `app/react/portainer/remote-portainers/queries/useRemotePortainer.ts`
- Create: `app/react/portainer/remote-portainers/queries/useRemotePortainerMutations.ts`
- Create: `app/react/portainer/remote-portainers/queries/useRemoteStacks.ts`
- Create: `app/react/portainer/remote-portainers/queries/useRemoteStackFile.ts`
- Create: `app/react/portainer/remote-portainers/remote-portainers.service.ts`
- Create: `app/react/portainer/remote-portainers/index.ts`

**Step 1: Define types**

Types:
- `RemotePortainer`
- `RemotePortainerCreatePayload`
- `RemotePortainerUpdatePayload`
- `RemoteStackUpdatePayload`
- `RemoteStackFileResponse`

**Step 2: Implement service**

Use the app's existing axios/api helper pattern from nearby `settings.service.ts` and environment query files.

Functions:
- `getRemotePortainers`
- `createRemotePortainer`
- `updateRemotePortainer`
- `deleteRemotePortainer`
- `testRemotePortainer`
- `getRemoteStacks`
- `getRemoteStackFile`
- `updateRemoteStack`

**Step 3: Implement React Query hooks**

Add list/detail queries and mutations with invalidation of remote Portainer and remote stack keys.

**Step 4: Test**

Add unit tests only if a clear local pattern exists for query hooks. Otherwise rely on TypeScript/build verification.

Run:

```bash
pnpm lint --filter app
pnpm test -- remote-portainers
```

If those commands do not match this repository's scripts, inspect `package.json` and use the nearest existing lint/test scripts.

**Step 5: Commit**

```bash
git add app/react/portainer/remote-portainers
git commit -m "local: add remote portainer frontend API layer"
```

## Task 5: Frontend Views

**Files:**
- Create: `app/react/portainer/remote-portainers/ListView/ListView.tsx`
- Create: `app/react/portainer/remote-portainers/ListView/RemotePortainersDatatable.tsx`
- Create: `app/react/portainer/remote-portainers/EditView/EditView.tsx`
- Create: `app/react/portainer/remote-portainers/StacksView/StacksView.tsx`
- Create: `app/react/portainer/remote-portainers/StackEditView/StackEditView.tsx`
- Create: `app/portainer/react/views/remote-portainers.ts`
- Modify: `app/portainer/react/views/index.ts`
- Modify: `app/react/sidebar/SettingsSidebar.tsx`

**Step 1: Add Angular state wrapper**

Follow the pattern in `app/portainer/react/views/environments.ts`.

Add states:

```text
portainer.remotePortainers
portainer.remotePortainers.new
portainer.remotePortainers.edit
portainer.remotePortainers.stacks
portainer.remotePortainers.stack
```

Register React components with `r2a(withUIRouter(withReactQuery(withCurrentUser(...))))`.

**Step 2: Add sidebar item**

In `app/react/sidebar/SettingsSidebar.tsx`, add an admin-only item under Environment-related or as a standalone Administration item:

```tsx
<SidebarItem
  label="Remote Portainers"
  to="portainer.remotePortainers"
  icon={Network}
  data-cy="portainerSidebar-remotePortainers"
/>
```

Use a lucide icon that already exists in the installed version.

**Step 3: Build connection list view**

List columns:
- Name
- URL
- TLS skip verify
- Updated
- Actions: Test, Edit, Stacks, Delete

Do not display API tokens.

**Step 4: Build create/edit view**

Fields:
- Name
- URL
- API token
- TLS skip verify

On edit, token field placeholder should indicate that leaving it empty keeps the existing token.

**Step 5: Build remote stacks view**

Show remote stack rows:
- Name
- Type
- EndpointId
- Status
- UpdatedBy/UpdateDate when available
- Action: Edit

Keep labels clear that these are remote Portainer stacks, not local stacks.

**Step 6: Build remote stack edit view**

Load remote stack file and show an editor.

Controls:
- environment variables fieldset if reuse is straightforward;
- prune switch;
- repull/redeploy switch;
- save button.

Submit to local `PUT /api/remote_portainers/{id}/stacks/{stackId}`.

**Step 7: Test**

Run the frontend type/lint/test commands available in `package.json`.

Expected: no TypeScript or lint failures.

**Step 8: Commit**

```bash
git add app/react/portainer/remote-portainers app/portainer/react/views/remote-portainers.ts app/portainer/react/views/index.ts app/react/sidebar/SettingsSidebar.tsx
git commit -m "local: add remote portainer administration UI"
```

## Task 6: End-to-End Manual Verification

**Files:**
- No required file changes.

**Step 1: Start a local dev Portainer**

Use the repository's documented dev command. If unknown, inspect `README.md`, `Makefile`, and `dev/`.

**Step 2: Prepare a remote Portainer**

Use either:
- an existing test Portainer with a manually created API token;
- or a second local Portainer container on a different port with its own data volume.

**Step 3: Verify connection flow**

Actions:
- create Remote Portainer connection;
- test connection;
- reload page;
- confirm token is not displayed;
- edit name/URL without changing token;
- test connection again.

Expected: connection remains usable.

**Step 4: Verify stack read flow**

Actions:
- open Remote Portainer stacks;
- open a stack file.

Expected: stack list and stack file match the remote Portainer UI.

**Step 5: Verify stack update flow**

Actions:
- edit a non-critical test stack;
- change image tag or env value;
- save with prune disabled;
- observe remote Portainer stack status.

Expected: the update is visible in the remote Portainer and the local Portainer has not created a local `Stack` record for it.

**Step 6: Commit verification notes only if needed**

If a manual test note is useful for the overlay branch, add it to `BRANCHES.md` on `local/meta`, not to upstream mirror branches.

## Task 7: Follow-Up Backlog

Do not implement these in MVP.

- Group two remote Portainers into an organization pair: primary and standby.
- Compare same-named stack files between primary and standby.
- Copy stack file from primary to standby with explicit confirmation.
- Add remote container/service read-only views.
- Add remote runtime actions: restart, stop, remove, service update.
- Add audit log entries for remote actions if local audit infrastructure is straightforward.
- Add token encryption if not implemented in MVP.

## Final Verification Before Merge Into Deploy

Run:

```bash
go test ./api/dataservices/remoteportainer ./api/remoteportainer/client ./api/http/handler/remoteportainers
go test ./api/http/handler
pnpm test -- remote-portainers
pnpm lint
```

Adjust frontend commands to actual scripts in `package.json`.

Expected: all selected backend and frontend checks pass.
