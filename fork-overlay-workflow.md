# Fork Overlay Workflow

This fork follows an overlay workflow. The goal is to keep upstream updates
simple while preserving local changes as small, named branches.

## Branch Roles

- `fork/develop` is the normal overlay base for this fork. It is the pinned
  upstream snapshot used to build local fork branches for the deployed product.
- Local `develop` is not the normal overlay base. Use it only for upstream PR
  work or when explicitly refreshing the fork's pinned upstream base.
- `local/<topic>` branches contain private fork changes. One branch equals one
  logical change.
- `deploy` is the runnable overlay branch. It is assembled from the pinned
  upstream base plus finished topical branches.
- `local/meta` contains repository-only documentation such as `AGENTS.md`,
  `fork-overlay-workflow.md`, and `BRANCHES.md`.

## Non-Negotiable Rules

1. Do not develop directly on `deploy`.
2. Do not use local `develop` for ordinary fork overlay work.
3. Do not create topical branches from `deploy`. `deploy` is never a base for
   new work, including temporary fixes.
4. Make every feature/fix in a separate topical branch first, based on
   `fork/develop` unless the user explicitly names another pinned base.
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
git fetch fork

git checkout -b local/<topic> fork/develop
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

# Only do this when intentionally refreshing the pinned fork base.
git checkout develop
git merge --ff-only upstream/<chosen-stable-branch>
git push fork develop

for branch in local/<topic-1> local/<topic-2>; do
  git checkout "$branch"
  git rebase fork/develop
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
fix, first create a `local/<topic>` branch from `fork/develop` or switch to the
existing `local/<topic>` branch that owns the feature.

If a change was accidentally made directly on `deploy`, do not continue building
on that mistake. Create a topical branch for the change, document it in
`BRANCHES.md`, and only then decide how to repair `deploy`.
