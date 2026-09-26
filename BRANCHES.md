# Active Overlay Branches

Base branch: `base/portainer-2.45.1` (upstream tag `2.45.1`, commit `bcfb8d279`).

Upstream release source: upstream tag `2.45.1`. Previous base: `base/portainer-2.45.0` (`upstream/release/2.45.0`), migrated 2026-09-26; pre-migration deploy backup: `backup/deploy-before-2.45.1-rebuild`.

`base/portainer-2.45.1` must point exactly at upstream tag `2.45.1`.
Ordinary fork overlay branches must be based on that release base.

Deployment branch: `deploy`.

Remote:

- `fork`: `https://github.com/InfoParkPro/portainer.git`
- `upstream`: `https://github.com/portainer/portainer.git` fetch-only

## Base Migration

Current status: `deploy` is assembled from `base/portainer-2.45.1` plus the
listed overlay branches.

## Branches

| Branch | Purpose | Current commit | Included in `deploy` |
|---|---|---:|---|
| `local/remote-portainer-api-token` | Adds Remote Portainer records managed with manually pasted API tokens; supports listing, editing, updating remote stacks through the remote Portainer API, and formats the remote list updated timestamp. | `887954c16` | Yes, equivalent deploy commits `fb0509b45`, `aea4a0157` |
| `local/hide-business-upsell` | Hides the "Upgrade to Business Edition" sidebar banner behind `PORTAINER_HIDE_BUSINESS_UPSELL=true` and brands CE UI surfaces as Portainer InfoPark Edition. | `1f4509e6b` | Yes, equivalent deploy commits `6b60464e3`, `fd5708d14` |
| `local/ci-deploy` | Adds GitHub Actions build for `deploy` and publishes only `ghcr.io/infoparkpro/portainer:latest`; final `deploy` assembly keeps the lockfile valid for the assembled 2.45 tree. | `59d8c38b8` | Yes, equivalent deploy commits `b4b256783`, `7a8e2c855`, `60645faa8`, `3f94e1e65` |
| `local/remote-portainer-create-form-fix` | Fixes Add Remote Portainer form staying in a loading state on new records. | `c612ec86b` | Yes, equivalent deploy commit `446cb1f41` |
| `local/service-task-actions` | Adds service task quick actions for container exec and force remove. | `c61e96beb` | Yes, equivalent deploy commit `453a50759` |
| `local/api-token-access-presets` | Adds API token presets: disabled, read-only, power, and manage. | `0e11dfe23` | Yes, equivalent deploy commit `47ad73ffa` |
| `local/api-token-temporary-elevation` | Adds temporary API token elevation with stored expiry, UI quick buttons for Manage access, request-time effective preset checks, `effectiveAccessPreset` in token API responses, and a current API-key inspection endpoint. | `765b5022a` | Yes, equivalent deploy commits `c4bae5223`, `8616ea72a`, `2e8972321` |
| `local/api-token-power-service-update` | Extends Power API-token preset to allow safe service force update through `PUT /api/endpoints/{id}/forceupdateservice` and restricted Docker exec for labelled safe containers, including websocket exec target rechecks and ResourceControl-independent create/start/resize after safety validation. Exec safety checks compare only host-side mount sources: container-internal mount targets are ignored, volume/tmpfs/image sources are daemon-managed, and denials log reasons at debug level. | `9c5d84f26` | Yes, equivalent deploy commits `eedad8572`, `f89166cab`, `457df999e`, `05dec6048`, `17b7bbf42`, `d3f217e79` |
| `local/hide-stack-migration-form` | Hides stack migration/duplication form by default behind a Migration button. | `44cba3749` | Yes, equivalent deploy commit `27b187a04` |
| `local/stack-services-refresh-button` | Adds a manual Refresh action to the services table so stack service/task status can be refreshed without enabling auto-refresh; refetches active stack service/task queries directly to avoid stale cached status. | `8e798c0af` | Yes, equivalent deploy commits `104b1d86c`, `a6720a728` |
| `local/keep-failed-stack-edits` | Keeps the edited Compose/Swarm stack file after an update deploy fails, so editor changes are not lost when image pull or deploy fails. | `7944882a4` | Yes, equivalent deploy commit `15ab21657` |
| `local/stack-webhooks` | Enables ordinary stack webhooks in the fork: file-based stacks can store webhook IDs, public webhook calls redeploy the saved stack with image pull, and each webhook is throttled to one accepted run per 10 minutes. | `f87006a28` | Yes, equivalent deploy commits `ca30e47e2`, `aaf72726e`, `12f1f630a`, `d88b8310d` |
| `local/self-update-helper` | Adds a Settings panel and backend helper mode for self-updating plain Docker Portainer containers; blocks Swarm service and Compose deployments; discovers the active container safely, prevents concurrent helpers, and preserves rollback state. | `8b8471f79` | Yes, equivalent deploy commits `742360dad`, `c18c35bdb` |
| `local/published-port-link-menu` | Replaces direct Published Ports links with a menu of current host, environment URL, and published host targets, each with copy, HTTP, and HTTPS actions. | `859f014e7` | Yes, equivalent deploy commit `c61f87dca` |
| `local/llms-capabilities` | Adds offline LLM discovery through `/llms.txt` (full HTTP-only API cookbook covering auth, Docker proxy operations, stacks, edge, kubernetes, registries, users, templates, webhooks, settings, and fork extras) and machine-readable fork capabilities through `/api/system/fork-capabilities`, serves the embedded OpenAPI 2.0 spec at `GET /api/docs/openapi.yaml`, including Power-token exec safety rules and current API-key discovery. | `805aab8b4` | Yes, equivalent deploy commits `5658fd59b`, `329a0a1c2`, `fb4b78f02`, `343b6e831` |
| `local/swarm-task-health` | Shows container health status for Swarm stack/service tasks when a related container is available, including ordinary Docker socket endpoints. | `53d165d97` | Yes, equivalent deploy commit `a7b4e6c73` |
| `local/docker-config-registry-auth-prune` | Rebuilds Docker CLI inline registry auths from the current Portainer registry list for each stack operation so stale auths from `/data/docker_config/config.json` are not reused after registries are deleted. | `9a7ca0414` | Yes, equivalent deploy commit `f05af9d85` |
| `local/browser-tab-title` | Updates the browser tab title from the current page header context and environment name, so stack/container pages are distinguishable across tabs. | `8d607b7aa` | Yes, equivalent deploy commits `fe7796d52`, `b6016d46c` |
| `local/auth-login-race-fix` | Retries the immediate post-login `/api/users/me` request once on `401` to tolerate browser/proxy cookie timing races without changing ordinary authenticated request handling; upstream 2.45.0 also includes `fix(login): handle same path RETURN_URL to avoid hung login`. | `74527dc27` | Yes, equivalent deploy commit `1d6d3881c` |
| `local/belief-map-gitignore` | Ignores local `.belief_map*` code-search artifacts so generated architecture maps can stay in the workspace without entering git status. | `d56442f80` | Yes, equivalent deploy commit `137360b0b` |
| `local/compose-on-swarm-create` | Lets Swarm Docker endpoints create either Swarm stacks or Compose stacks from the create-stack UI; Compose mode uses the standalone stack API while keeping Swarm as the default. | `968217a87` | Yes, equivalent deploy commit `ce17b46fe` |
| `local/fix-current-password-autocomplete` | Marks the access-token confirmation password as the current password so browser password managers can autofill it. | `b5d3c9de0` | Yes, equivalent deploy commit `6d70f12ee` |
| `local/fix-image-export-query-params` | Serializes image export names as repeated Docker API query parameters so exported archives contain the selected images. | `f279fe50a` | Yes, equivalent deploy commit `cd6486ca0` |
| `local/improve-log-viewer-ui` | Improves Docker log viewing with collapsible settings, quick filtering, copy/download actions, horizontal scrolling, and responsive heights. | `069ed09fc` | Yes, equivalent deploy commit `df40c8947` |
| `local/audit-log` | Enables the CE activity log UI and adds a local audit backend: stack create/update/delete/git redeploy, public stack webhook redeploy, service force update, and docker image pull/import are recorded with endpoint context, acting user, and API key attribution; admin-only list and CSV API at `/api/useractivity/logs`. | `e6972722a` | Yes, equivalent deploy commit `c95de6b1b` |
| `local/meta` | Repository workflow docs: `AGENTS.md`, `fork-overlay-workflow.md`, `BRANCHES.md`. | `HEAD` | No |

## Obsolete Branches

| Branch | Status |
|---|---|
| `local/remote-portainer-date-format` | Merged into `local/remote-portainer-api-token` as `887954c16`; can be deleted after push. |

## Deploy Contents

`deploy` currently includes:

- Base Portainer release `2.45.1` with fork version bump to `2.45.9`.
- Upstream login fix for same-document `RETURN_URL` hung login.
- Remote Portainer API-token management.
- Hidden Business upsell banner.
- Portainer InfoPark Edition branding on CE footer/About surfaces.
- GHCR latest-only image build.
- pnpm lockfile valid for the assembled 2.45 deploy tree.
- Webhook builder tests adapted to the 2.45 `setGeneralInfo` signature (`f661d4218`).
- `.pnpm-store/` gitignore entry (`f989b43c6`).
- Add Remote Portainer create-form loading fix.
- Service task exec/remove actions.
- API token access presets.
- Power API-token service force update.
- Restricted Power API-token Docker exec for labelled safe containers, with
  create/start/resize independent of ResourceControl after safety validation.
  Safety checks compare only host-side mount sources (container-internal
  targets and daemon-managed volume sources are ignored) and denials carry
  reasons in debug logs and websocket errors. The preset whitelist matches
  the stripped proxy path shapes `/{id}/docker/...` and
  `/{id}/agent/docker/...` that the middleware sees after the
  `/api/endpoints` prefix strip, in addition to the portainer API form.
- Remote Portainer updated-date formatting.
- Hidden stack migration form.
- Manual stack services refresh action with direct refetch of active resource
  queries.
- Failed Compose/Swarm stack deploys keep the edited stack file instead of
  rolling it back.
- Ordinary stack webhooks with a 10 minute throttle.
- Portainer self-update helper for plain Docker containers with active-container
  discovery, concurrent-update locking, and restart-safe rollback handling.
- Published Ports menu with copy, HTTP, and HTTPS actions.
- Offline LLM discovery through `/llms.txt` (full HTTP-only API cookbook),
  `/api/system/fork-capabilities`, the embedded OpenAPI 2.0 spec at
  `GET /api/docs/openapi.yaml`, and current API-key discovery.
- Swarm task container health status in stack/service task tables.
- Pruning stale Docker CLI inline registry auths during stack operations.
- Browser tab titles with page context and environment name.
- Post-login current-user retry for intermittent `/api/users/me` 401 races.
- Ignored local `.belief_map*` code-search artifacts.
- Compose-stack creation mode on Swarm Docker endpoints.
- Current-password autofill for access-token confirmation.
- Temporary API token elevation with stored expiry, Manage quick buttons,
  `effectiveAccessPreset` in token API responses, and
  `/api/users/me/current-api-key`.
- Correct Docker image export query serialization.
- Improved Docker log viewer controls and responsive layout.
- Activity log with audit records for stack operations, service force update,
  and image pulls, including user and API key attribution, and admin-only
  list/CSV API at `/api/useractivity/logs` (UI enabled in CE).

## Operating Notes

- Do not push without explicit user instruction.
- Do not add new work directly to `deploy`.
- During approved deploy assembly, update deterministic final-tree build
  artifacts such as lockfile checksums directly in `deploy`; do not create a
  standalone topical branch only for those files.
- When a branch is added to `deploy`, update this file in `local/meta`.
- For upstream Portainer PR work only, fetch `upstream/develop` temporarily,
  create a short-lived `upstream-pr/<topic>` branch, submit the PR, then delete
  the temporary branch. Do not use it as an overlay base.
