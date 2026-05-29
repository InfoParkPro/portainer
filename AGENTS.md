# Repository Agent Instructions

This repository is a fork overlay. Follow `fork-overlay-workflow.md` in this
repository before making code or documentation changes.

## Hard Rules

- Do not develop directly on `deploy`.
- Do not edit files, commit, or run implementation work while checked out on
  `deploy`, except when explicitly applying already-finished branch commits into
  `deploy`.
- Every change starts in a separate topical branch from `develop` unless the
  user explicitly gives a different base. Use one branch per logical change.
- Use `local/<topic>` for fork-only/private changes.
- `deploy` is an overlay branch only: it is assembled from `develop` plus
  finished topical branches by cherry-pick or merge.
- When the user asks to add a feature/fix to `deploy`, first make and verify the
  change in its own branch, then apply the finished commit to `deploy`.
- Never push any branch, including `deploy`, unless the user explicitly says to
  push.
- If currently on `deploy` and asked to change something, switch to or create a
  topical branch before editing.
- If `deploy` has local commits ahead of the remote, preserve them. Do not
  reset, rebase, squash, or drop them unless the user explicitly asks.

## Required Files

- `fork-overlay-workflow.md` explains the local workflow.
- `BRANCHES.md` records active overlay branches and what each branch adds.

