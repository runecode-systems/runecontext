#!/usr/bin/env bash

set -euo pipefail

if [[ "${#}" -lt 1 ]]; then
  printf '%s\n' "usage: validate_after_authoritative_edit.sh <changed-path>..." >&2
  exit 2
fi

find_repo_root() {
  local start
  start="$(pwd)"
  while true; do
    if [[ -f "$start/runecontext.yaml" ]]; then
      printf '%s\n' "$start"
      return 0
    fi
    if [[ "$start" == "/" ]]; then
      return 1
    fi
    start="$(dirname "$start")"
  done
}

is_authored_authoritative_path() {
  local rel
  rel="$1"

  case "$rel" in
    runecontext.yaml) return 0 ;;
    runecontext/bundles/*.yaml) return 0 ;;
    runecontext/project/*.md) return 0 ;;
    runecontext/standards/*.md|runecontext/standards/**/*.md) return 0 ;;
    runecontext/specs/*.md|runecontext/specs/**/*.md) return 0 ;;
    runecontext/decisions/*.md|runecontext/decisions/**/*.md) return 0 ;;
    runecontext/changes/*/status.yaml) return 0 ;;
    runecontext/changes/*/proposal.md) return 0 ;;
    runecontext/changes/*/standards.md) return 0 ;;
    runecontext/changes/*/design.md) return 0 ;;
    runecontext/changes/*/verification.md) return 0 ;;
    runecontext/changes/*/tasks.md) return 0 ;;
    runecontext/changes/*/references.md) return 0 ;;
    .runecontext/*) return 1 ;;
    *) return 1 ;;
  esac
}

repo_root=""
if repo_root="$(find_repo_root)"; then
  :
else
  # Outside RuneContext projects, fail soft by doing nothing.
  exit 0
fi

should_validate=0
for changed_path in "$@"; do
  # Normalize to a repo-root-relative slash path when possible.
  if [[ "$changed_path" = /* ]]; then
    rel="${changed_path#"$repo_root"/}"
    if [[ "$rel" == "$changed_path" ]]; then
      continue
    fi
  else
    rel="$changed_path"
  fi
  rel="${rel#./}"
  if is_authored_authoritative_path "$rel"; then
    should_validate=1
    break
  fi
done

if [[ "$should_validate" -eq 1 ]]; then
  exec runectx validate --path "$repo_root"
fi

exit 0
