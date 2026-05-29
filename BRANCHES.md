# Active Overlay Branches

Base branch: `develop`.

Deployment branch: `deploy`.

Remote:

- `fork`: `https://github.com/InfoParkPro/portainer.git`
- `upstream`: `https://github.com/portainer/portainer.git` fetch-only

## Branches

| Branch | Purpose | Current commit | Included in `deploy` |
|---|---|---:|---|
| `local/remote-portainer-api-token` | Adds Remote Portainer records managed with manually pasted API tokens; supports listing, editing, updating remote stacks through the remote Portainer API, and formats the remote list updated timestamp. | `887954c16` | Yes |
| `local/hide-business-upsell` | Hides the "Upgrade to Business Edition" sidebar banner behind `PORTAINER_HIDE_BUSINESS_UPSELL=true`. | `ac98a3a7b` | Yes |
| `local/ci-deploy` | Adds GitHub Actions build for `deploy` and publishes only `ghcr.io/infoparkpro/portainer:latest`. | `59d8c38b8` | Yes |
| `local/remote-portainer-create-form-fix` | Fixes Add Remote Portainer form staying in a loading state on new records. | `c612ec86b` | Yes |
| `local/service-task-actions` | Adds service task quick actions for container exec and force remove. | `c61e96beb` | Yes |
| `local/api-token-access-presets` | Adds API token presets: disabled, read-only, power, and manage. | `0e11dfe23` | Yes |
| `local/hide-stack-migration-form` | Hides stack migration/duplication form by default behind a Migration button. | `44cba3749` | Yes, equivalent deploy commit `421324b0b` |
| `local/meta` | Repository workflow docs: `AGENTS.md`, `fork-overlay-workflow.md`, `BRANCHES.md`. | `2d5d3c1c2` | No |

## Obsolete Branches

| Branch | Status |
|---|---|
| `local/remote-portainer-date-format` | Merged into `local/remote-portainer-api-token` as `887954c16`; can be deleted after push. |

## Deploy Contents

`deploy` currently includes:

- Remote Portainer API-token management.
- Hidden Business upsell banner.
- GHCR latest-only image build.
- Add Remote Portainer create-form loading fix.
- Service task exec/remove actions.
- API token access presets.
- Remote Portainer updated-date formatting.
- Hidden stack migration form.

## Operating Notes

- Do not push without explicit user instruction.
- Do not add new work directly to `deploy`.
- When a branch is added to `deploy`, update this file in `local/meta`.
