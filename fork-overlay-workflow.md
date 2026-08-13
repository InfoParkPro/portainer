# Fork Overlay Workflow

This fork follows an overlay workflow. The goal is to keep upstream updates
simple while preserving local changes as small, named branches.

## Branch Roles

- `base/portainer-2.44.0` is the only normal overlay base for this fork. It must
  point exactly at `upstream/release/2.44.0`.
- `local/<topic>` branches contain private fork changes. One branch equals one
  logical change.
- `deploy` is the runnable overlay branch. It is assembled from the pinned
  upstream base plus finished topical branches.
- `local/meta` contains repository-only documentation such as `AGENTS.md`,
  `fork-overlay-workflow.md`, and `BRANCHES.md`.

## Non-Negotiable Rules

1. Do not work directly on `deploy`.
2. Do not use any non-release upstream branch for ordinary fork overlay work.
3. Do not create topical branches from `deploy`. `deploy` is never a base for
   new work, including temporary fixes.
4. Make every feature/fix in a separate topical branch first, based on
   `base/portainer-2.44.0` unless the user explicitly names another release
   base.
5. Fixes to an existing overlay feature go into that feature's existing
   `local/<topic>` branch, then get re-applied to `deploy`.
6. Verify the topical branch before applying it to `deploy`.
7. Apply finished work to `deploy` only by cherry-pick or merge.
8. Push only after the user explicitly asks for a push.
9. Never reset, rebase, squash, or drop existing `deploy` commits unless the user
   explicitly asks.
10. Keep `BRANCHES.md` updated when adding, removing, or changing overlay
   branches.
11. Do not create a topical branch only for integration artifacts that are
    produced by assembling `deploy`, such as `pnpm-lock.yaml` checksum updates.
    Those fixes belong directly in `deploy` during the approved deploy assembly.

## Normal Change Flow

```bash
git fetch upstream

git checkout -b local/<topic> base/portainer-2.44.0
# edit, test, commit

# only when the user asks to include the finished change in deploy:
git checkout deploy
git cherry-pick <topic-commit>

# only when the user explicitly says to push:
git push fork deploy
```

## Updating From Upstream

```bash
git fetch upstream

git branch -f base/portainer-2.44.0 upstream/release/2.44.0

for branch in local/<topic-1> local/<topic-2>; do
  git checkout "$branch"
  git rebase base/portainer-2.44.0
done

git checkout deploy
# Re-apply the verified local branches according to BRANCHES.md.
```

If upstream uses a newer release branch for the current Portainer version,
document the chosen release base in `BRANCHES.md` before rebuilding `deploy`.

## Upstream PR Work

Ordinary fork overlay work must not use `upstream/develop`.

When preparing a PR for `portainer/portainer`, fetch `upstream/develop` only
for that PR, create a short-lived `upstream-pr/<topic>` branch, submit the PR,
then delete the temporary branch. Do not record that branch as a fork overlay
base and do not use it for `deploy`.

## Deploy Branch Policy

`deploy` is not a work branch. It is allowed to contain commits that apply local
branches, but those commits must already exist and be verified on their own
topical branches.

There is one integration exception: while assembling `deploy` from the pinned
release base and verified topical branches, files whose correct content depends
on the final assembled tree may be fixed directly on `deploy`. Examples include
lockfiles, generated checksums, or other deterministic build metadata. Keep
these commits narrowly scoped, verify the failing build step, and record them in
`BRANCHES.md`. Do not create a `local/<topic>` branch only for such generated
deploy integration files.

Never run `git checkout -b`, `git switch -c`, or any equivalent branch-creation
command while checked out on `deploy`. If you are on `deploy` and need to make a
fix, first create a `local/<topic>` branch from `base/portainer-2.44.0` or
switch to the existing `local/<topic>` branch that owns the feature.

If a change was accidentally made directly on `deploy`, do not continue building
on that mistake. Create a topical branch for the change, document it in
`BRANCHES.md`, and only then decide how to repair `deploy`.
