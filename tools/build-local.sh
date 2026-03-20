#!/usr/bin/env bash
set -euo pipefail

repo_root="$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
package_root="${repo_root}/build/local/runecontext"
binary_path="${package_root}/bin/runectx"
launcher_path="${repo_root}/bin/runectx"

entries=(
  README.md
  LICENSE
  NOTICE
  DCO
  CONTRIBUTING.md
  SECURITY.md
  CODE_OF_CONDUCT.md
  go.mod
  go.sum
  flake.nix
  flake.lock
  justfile
  docs
  core
  adapters
  schemas
  fixtures
  cmd
  internal
  tools
  nix
)

rm -rf "${package_root}"
mkdir -p "${package_root}/bin"

(
  cd "${repo_root}"
  go build -o "${binary_path}" ./cmd/runectx
)

for entry in "${entries[@]}"; do
  cp -R "${repo_root}/${entry}" "${package_root}/"
done

mkdir -p "${repo_root}/bin"
cat > "${launcher_path}" <<'EOF'
#!/usr/bin/env sh
set -eu

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
repo_root=$(CDPATH= cd -- "${script_dir}/.." && pwd)
target="${repo_root}/build/local/runecontext/bin/runectx"

if [ ! -x "${target}" ]; then
  printf 'missing local dogfood build at %s\n' "${target}" >&2
  printf 'run `just build` from %s first\n' "${repo_root}" >&2
  exit 1
fi

exec "${target}" "$@"
EOF
chmod +x "${launcher_path}"

printf 'Built local RuneContext package at %s\n' "${package_root}"
printf 'Use %s to dogfood this checkout\n' "${repo_root}/bin/runectx"
