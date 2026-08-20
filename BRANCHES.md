# Active Overlay Branches

Base branch: `base/portainer-2.44.0`.

Upstream release source: `upstream/release/2.44.0`.

`base/portainer-2.44.0` must point exactly at `upstream/release/2.44.0`.
Ordinary fork overlay branches must be based on that release base.

Deployment branch: `deploy`.

Remote:

- `fork`: `https://github.com/InfoParkPro/portainer.git`
- `upstream`: `https://github.com/portainer/portainer.git` fetch-only

## Base Migration

Current status: `deploy` is assembled from `base/portainer-2.44.0` plus the
listed overlay branches.

## Branches

| Branch | Purpose | Current commit | Included in `deploy` |
|---|---|---:|---|
| `local/remote-portainer-api-token` | Adds Remote Portainer records managed with manually pasted API tokens; supports listing, editing, updating remote stacks through the remote Portainer API, and formats the remote list updated timestamp. | `887954c16` | Yes, equivalent deploy commits `c14ba290b`, `85f1a7475` |
| `local/hide-business-upsell` | Hides the "Upgrade to Business Edition" sidebar banner behind `PORTAINER_HIDE_BUSINESS_UPSELL=true`. | `ac98a3a7b` | Yes, equivalent deploy commit `4dfb38ee0` |
| `local/ci-deploy` | Adds GitHub Actions build for `deploy` and publishes only `ghcr.io/infoparkpro/portainer:latest`; final `deploy` assembly also refreshes the pnpm lockfile checksum for the assembled 2.44 tree. | `59d8c38b8` | Yes, equivalent deploy commits `f5641afb9`, `b27a101e7`, `6af1347a5`, `bd94efb46` |
| `local/remote-portainer-create-form-fix` | Fixes Add Remote Portainer form staying in a loading state on new records. | `c612ec86b` | Yes, equivalent deploy commit `08ce62f05` |
| `local/service-task-actions` | Adds service task quick actions for container exec and force remove. | `c61e96beb` | Yes, equivalent deploy commit `183366a8d` |
| `local/api-token-access-presets` | Adds API token presets: disabled, read-only, power, and manage. | `0e11dfe23` | Yes, equivalent deploy commit `bd01c3656` |
| `local/api-token-power-service-update` | Extends Power API-token preset to allow safe service force update through `PUT /api/endpoints/{id}/forceupdateservice`. | `cf829fa7d` | Yes, equivalent deploy commit `fd20706ff` |
| `local/hide-stack-migration-form` | Hides stack migration/duplication form by default behind a Migration button. | `44cba3749` | Yes, equivalent deploy commit `19660d978` |
| `local/stack-services-refresh-button` | Adds a manual Refresh action to the services table so stack service/task status can be refreshed without enabling auto-refresh; refetches active stack service/task queries directly to avoid stale cached status. | `8e798c0af` | Yes, equivalent deploy commits `caf01769d`, `84fa0450d` |
| `local/keep-failed-stack-edits` | Keeps the edited Compose/Swarm stack file after an update deploy fails, so editor changes are not lost when image pull or deploy fails. | `7944882a4` | Yes, equivalent deploy commit `e09c4eb11` |
| `local/stack-webhooks` | Enables ordinary stack webhooks in the fork: file-based stacks can store webhook IDs, public webhook calls redeploy the saved stack with image pull, and each webhook is throttled to one accepted run per 10 minutes. | `f87006a28` | Yes, equivalent deploy commits `56b4e3bed`, `81ad44b02`, `aae54e629`, `319068d1e` |
| `local/self-update-helper` | Adds a Settings panel and backend helper mode for self-updating plain Docker Portainer containers; blocks Swarm service and Compose deployments. | `d50b5264b` | Yes, equivalent deploy commit `f1862a71b` |
| `local/published-port-link-menu` | Replaces direct Published Ports links with a menu of current host, environment URL, and published host targets, each with copy, HTTP, and HTTPS actions. | `859f014e7` | Yes, equivalent deploy commit `a697400c0` |
| `local/llms-capabilities` | Adds offline LLM discovery through `/llms.txt` and machine-readable fork capabilities through `/api/system/fork-capabilities`. | `ec9e0a36a` | Yes, equivalent deploy commit `4b88ce308` |
| `local/swarm-task-health` | Shows container health status for Swarm stack/service tasks when a related container is available, including ordinary Docker socket endpoints. | `53d165d97` | Yes, equivalent deploy commit `099afb696` |
| `local/docker-config-registry-auth-prune` | Rebuilds Docker CLI inline registry auths from the current Portainer registry list for each stack operation so stale auths from `/data/docker_config/config.json` are not reused after registries are deleted. | `9a7ca0414` | Yes, equivalent deploy commit `ecc024cf8` |
| `local/browser-tab-title` | Updates the browser tab title from the current page header context and environment name, so stack/container pages are distinguishable across tabs. | `8d607b7aa` | Yes, equivalent deploy commits `5f86b9dc3`, `e7d4f01e6` |
| `local/auth-login-race-fix` | Retries the immediate post-login `/api/users/me` request once on `401` to tolerate browser/proxy cookie timing races without changing ordinary authenticated request handling. | `74527dc27` | Yes, equivalent deploy commit `b48b2a8ae` |
| `local/belief-map-gitignore` | Ignores local `.belief_map*` code-search artifacts so generated architecture maps can stay in the workspace without entering git status. | `d56442f80` | Yes, equivalent deploy commit `949081249` |
| `local/meta` | Repository workflow docs: `AGENTS.md`, `fork-overlay-workflow.md`, `BRANCHES.md`. | `HEAD` | No |

## Obsolete Branches

| Branch | Status |
|---|---|
| `local/remote-portainer-date-format` | Merged into `local/remote-portainer-api-token` as `887954c16`; can be deleted after push. |

## Deploy Contents

`deploy` currently includes:

- Version bump to `2.44.1`.
- Remote Portainer API-token management.
- Hidden Business upsell banner.
- GHCR latest-only image build.
- pnpm lockfile checksum refreshed for the assembled 2.44 deploy tree.
- Add Remote Portainer create-form loading fix.
- Service task exec/remove actions.
- API token access presets.
- Power API-token service force update.
- Remote Portainer updated-date formatting.
- Hidden stack migration form.
- Manual stack services refresh action with direct refetch of active resource
  queries.
- Failed Compose/Swarm stack deploys keep the edited stack file instead of
  rolling it back.
- Ordinary stack webhooks with a 10 minute throttle.
- Portainer self-update helper for plain Docker containers.
- Published Ports menu with copy, HTTP, and HTTPS actions.
- Offline LLM discovery through `/llms.txt` and `/api/system/fork-capabilities`.
- Swarm task container health status in stack/service task tables.
- Pruning stale Docker CLI inline registry auths during stack operations.
- Browser tab titles with page context and environment name.
- Post-login current-user retry for intermittent `/api/users/me` 401 races.
- Ignored local `.belief_map*` code-search artifacts.

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
