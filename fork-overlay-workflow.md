# Fork Overlay Workflow

This fork follows an overlay workflow. The goal is to keep upstream updates
simple while preserving local changes as small, named branches.

## Branch Roles

- `develop` is the upstream mirror branch for this fork.
- `local/<topic>` branches contain private fork changes. One branch equals one
  logical change.
- `deploy` is the runnable overlay branch. It is assembled from `develop` plus
  finished topical branches.
- `local/meta` contains repository-only documentation such as `AGENTS.md`,
  `fork-overlay-workflow.md`, and `BRANCHES.md`.

## Non-Negotiable Rules

1. Do not develop directly on `deploy`.
2. Do not commit code directly on `develop`.
3. Do not create topical branches from `deploy`. `deploy` is never a base for
   new work, including temporary fixes.
4. Make every feature/fix in a separate topical branch first.
5. Fixes to an existing overlay feature go into that feature's existing
   `local/<topic>` branch, then get re-applied to `deploy`.
6. Verify the topical branch before applying it to `deploy`.
7. Apply finished work to `deploy` only by cherry-pick or merge.
8. Push only after the user explicitly asks for a push.
9. Never reset, rebase, squash, or drop existing `deploy` commits unless the user
   explicitly asks.
10. Keep `BRANCHES.md` updated when adding, removing, or changing overlay
   branches.

## Normal Change Flow

```bash
git checkout develop
git pull --ff-only fork develop

git checkout -b local/<topic>
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

git checkout develop
git merge --ff-only upstream/develop

for branch in local/<topic-1> local/<topic-2>; do
  git checkout "$branch"
  git rebase develop
done

git checkout deploy
# Re-apply the verified local branches according to BRANCHES.md.
```

If upstream uses a different stable branch for the current Portainer version,
document the chosen upstream base in `BRANCHES.md`.

## Deploy Branch Policy

`deploy` is not a development branch. It is allowed to contain commits that
apply local branches, but those commits must already exist and be verified on
their own topical branches.

Never run `git checkout -b`, `git switch -c`, or any equivalent branch-creation
command while checked out on `deploy`. If you are on `deploy` and need to make a
fix, first switch to `develop` or to the existing `local/<topic>` branch that
owns the feature.

If a change was accidentally made directly on `deploy`, do not continue building
on that mistake. Create a topical branch for the change, document it in
`BRANCHES.md`, and only then decide how to repair `deploy`.
