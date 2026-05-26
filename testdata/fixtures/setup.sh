#!/usr/bin/env bash
# Creates a deterministic fixture git repository.
# Idempotent — safe to re-run; always produces the same commit hashes.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO="$SCRIPT_DIR/repo"

# Pinned identity + timestamps → deterministic hashes across machines.
export GIT_AUTHOR_NAME="Fixture Author"
export GIT_AUTHOR_EMAIL="fixture@example.com"
export GIT_COMMITTER_NAME="Fixture Author"
export GIT_COMMITTER_EMAIL="fixture@example.com"

# ── clean slate ───────────────────────────────────────────────────────────────
rm -rf "$REPO"
mkdir -p "$REPO"
cd "$REPO"

git init
git symbolic-ref HEAD refs/heads/main
git config user.email ${GIT_AUTHOR_EMAIL}
git config user.name ${GIT_AUTHOR_NAME}

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
git checkout -b feature/my-feature HEAD~1
echo "feature" > feature.txt
git add feature.txt
GIT_AUTHOR_DATE="2000-01-04T00:00:00Z" GIT_COMMITTER_DATE="2000-01-04T00:00:00Z" \
  git commit -m "add feature.txt"

# ── back to main ─────────────────────────────────────────────────────────────
git checkout main
