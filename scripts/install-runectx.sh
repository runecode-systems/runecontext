#!/usr/bin/env bash
set -euo pipefail

repo="runecode-systems/runecontext"
install_dir_default="${HOME}/.local/bin"

usage() {
  cat <<'EOF'
Install runectx from GitHub Releases.

Usage:
  install-runectx.sh [--version TAG] [--install-dir DIR] [--yes]

Options:
  --version TAG      Install a specific release tag (e.g., v0.1.0-alpha.8)
                     Defaults to the latest published release.
  --install-dir DIR  Install directory for runectx (default: $HOME/.local/bin)
  --yes              Skip confirmation prompt and continue install.
  --help             Show this help text.

Environment:
  RUNECTX_VERSION      Same as --version
  RUNECTX_INSTALL_DIR  Same as --install-dir
EOF
}

require_cmd() {
  local cmd="$1"
  if ! command -v "${cmd}" >/dev/null 2>&1; then
    printf 'missing required command: %s\n' "${cmd}" >&2
    exit 1
  fi
}

resolve_latest_tag() {
  local latest_url
  latest_url="$(curl -fsSL -o /dev/null -w '%{url_effective}' "https://github.com/${repo}/releases/latest")"
  if [[ "${latest_url}" != *"/releases/tag/"* ]]; then
    printf 'failed to resolve latest release tag from %s\n' "${latest_url}" >&2
    exit 1
  fi
  printf '%s\n' "${latest_url##*/}"
}

validate_version_tag() {
  local tag="$1"
  if [[ "${tag}" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[A-Za-z0-9.-]+)?$ ]]; then
    return
  fi
  printf 'invalid release tag %q (expected format like v0.1.0-alpha.8)\n' "${tag}" >&2
  exit 1
}

map_os() {
  local uname_s
  uname_s="$(uname -s | tr '[:upper:]' '[:lower:]')"
  case "${uname_s}" in
    linux)
      printf 'linux\n'
      ;;
    darwin)
      printf 'darwin\n'
      ;;
    *)
      printf 'unsupported operating system: %s\n' "${uname_s}" >&2
      exit 1
      ;;
  esac
}

map_arch() {
  local uname_m
  uname_m="$(uname -m)"
  case "${uname_m}" in
    x86_64)
      printf 'amd64\n'
      ;;
    arm64|aarch64)
      printf 'arm64\n'
      ;;
    *)
      printf 'unsupported architecture: %s\n' "${uname_m}" >&2
      exit 1
      ;;
  esac
}

sha256_file() {
  local file_path="$1"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "${file_path}" | cut -d ' ' -f 1
    return
  fi
  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "${file_path}" | cut -d ' ' -f 1
    return
  fi
  printf 'missing checksum tool: need sha256sum or shasum\n' >&2
  exit 1
}

version="${RUNECTX_VERSION:-}"
install_dir="${RUNECTX_INSTALL_DIR:-${install_dir_default}}"
assume_yes=false

while [ "$#" -gt 0 ]; do
  case "$1" in
    --version)
      shift
      if [ "$#" -eq 0 ]; then
        printf '--version requires a value\n' >&2
        exit 1
      fi
      version="$1"
      ;;
    --install-dir)
      shift
      if [ "$#" -eq 0 ]; then
        printf '--install-dir requires a value\n' >&2
        exit 1
      fi
      install_dir="$1"
      ;;
    --yes)
      assume_yes=true
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      printf 'unknown argument: %s\n' "$1" >&2
      usage >&2
      exit 1
      ;;
  esac
  shift
done

require_cmd curl
require_cmd tar
require_cmd install
require_cmd grep

if [ -z "${version}" ]; then
  version="$(resolve_latest_tag)"
fi
validate_version_tag "${version}"

os="$(map_os)"
arch="$(map_arch)"

archive="runecontext_${version}_${os}_${arch}.tar.gz"
checksums="SHA256SUMS"
base_url="https://github.com/${repo}/releases/download/${version}"

workdir="$(mktemp -d)"
trap 'rm -rf "${workdir}"' EXIT

printf 'Resolving release: %s\n' "${version}"
printf 'Target platform: %s/%s\n' "${os}" "${arch}"
printf 'Install destination: %s/runectx\n' "${install_dir}"

curl -fsSL -o "${workdir}/${archive}" "${base_url}/${archive}"
curl -fsSL -o "${workdir}/${checksums}" "${base_url}/${checksums}"

expected_line="$(grep -F "  ${archive}" "${workdir}/${checksums}" || true)"
if [ -z "${expected_line}" ]; then
  printf 'SHA256SUMS does not contain an entry for %s\n' "${archive}" >&2
  exit 1
fi

expected_hash="${expected_line%% *}"
actual_hash="$(sha256_file "${workdir}/${archive}")"

printf '\nChecksum verification:\n'
printf '  archive:  %s\n' "${archive}"
printf '  expected: %s\n' "${expected_hash}"
printf '  actual:   %s\n' "${actual_hash}"

if [ "${actual_hash}" != "${expected_hash}" ]; then
  printf 'checksum verification failed\n' >&2
  exit 1
fi

printf '  result:   OK\n\n'

if [ "${assume_yes}" != true ]; then
  printf 'Continue with installation? [y/N]: '
  read -r reply
  case "${reply:-n}" in
    y|Y|yes|YES)
      ;;
    *)
      printf 'Installation cancelled.\n'
      exit 0
      ;;
  esac
fi

mkdir -p "${workdir}/unpack"
tar -xzf "${workdir}/${archive}" -C "${workdir}/unpack"

package_dir="${workdir}/unpack/runecontext_${version}_${os}_${arch}"
binary_path="${package_dir}/bin/runectx"

if [ ! -f "${binary_path}" ]; then
  printf 'expected binary not found: %s\n' "${binary_path}" >&2
  exit 1
fi

mkdir -p "${install_dir}"
install -m 0755 "${binary_path}" "${install_dir}/runectx"

printf '\nInstalled runectx to %s/runectx\n' "${install_dir}"
"${install_dir}/runectx" version

cat <<'EOF'

Next steps:
- Ensure your install directory is on PATH.
- Run: runectx doctor --path /path/to/project
- Initialize a project: runectx init --path /path/to/project
- Sync adapter files: runectx adapter sync --path /path/to/project <tool>
- Preview upgrades: runectx upgrade --path /path/to/project
EOF
