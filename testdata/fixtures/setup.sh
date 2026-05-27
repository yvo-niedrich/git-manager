#!/usr/bin/env bash
# Creates a deterministic fixture git repository.
# Concurrent-safe: builds in a sibling temp dir on the same filesystem then
# atomically renames into place. If two test binaries race, one wins; the
# other discards its work tree and exits cleanly.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO="$SCRIPT_DIR/repo"

# Fast path: already built by an earlier process in this test run.
if [ -d "$REPO/.git" ]; then
    exit 0
fi

# Pinned identity + timestamps → deterministic hashes across machines.
export GIT_AUTHOR_NAME="Fixture Author"
export GIT_AUTHOR_EMAIL="fixture@example.com"
export GIT_COMMITTER_NAME="Fixture Author"
export GIT_COMMITTER_EMAIL="fixture@example.com"

# Build in a sibling temp dir so we never delete the directory another process
# may already be sitting inside.
WORK=$(mktemp -d "$SCRIPT_DIR/.repo.XXXXXX")
trap 'rm -rf "$WORK"' EXIT

cd "$WORK"

git init -q
git symbolic-ref HEAD refs/heads/main
git config user.email "${GIT_AUTHOR_EMAIL}"
git config user.name "${GIT_AUTHOR_NAME}"

# ── main: three commits ───────────────────────────────────────────────────────
GIT_AUTHOR_DATE="2000-01-01T00:00:00Z" GIT_COMMITTER_DATE="2000-01-01T00:00:00Z" \
  git commit --allow-empty -m "initial commit"

echo "hello" > hello.txt
git add hello.txt
GIT_AUTHOR_DATE="2000-01-02T00:00:00Z" GIT_COMMITTER_DATE="2000-01-02T00:00:00Z" \
  git commit -m "second commit

This commit has a body paragraph to exercise multi-line message parsing."

git tag v0.1.0

echo "world" >> hello.txt
git add hello.txt
GIT_AUTHOR_DATE="2000-01-03T00:00:00Z" GIT_COMMITTER_DATE="2000-01-03T00:00:00Z" \
  git commit -m "third commit"

# ── simulate origin/main pointing at current HEAD ────────────────────────────
# A real remote config (with fetch refspec) is required for %(upstream:short)
# to resolve correctly in git branch --format.
git remote add origin /dev/null
git config remote.origin.fetch "+refs/heads/*:refs/remotes/origin/*"
git update-ref refs/remotes/origin/main HEAD
git branch --set-upstream-to=origin/main main

# ── feature branch off second commit ─────────────────────────────────────────
git checkout -q -b feature/my-feature HEAD~1
echo "feature" > feature.txt
git add feature.txt
GIT_AUTHOR_DATE="2000-01-04T00:00:00Z" GIT_COMMITTER_DATE="2000-01-04T00:00:00Z" \
  git commit -m "add feature.txt"

# ── back to main ─────────────────────────────────────────────────────────────
git checkout -q main

# Atomic rename into place. If another process already won the race, discard
# our work tree and exit cleanly — the repo is already there.
mv "$WORK" "$REPO" 2>/dev/null && trap - EXIT || true
