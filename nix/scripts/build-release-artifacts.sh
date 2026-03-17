#!/usr/bin/env bash
set -euo pipefail

umask 022

export CGO_ENABLED=0
export GOFLAGS="-trimpath -mod=vendor"
export SOURCE_DATE_EPOCH=315532800
export TZ=UTC
export LC_ALL=C

mkdir -p release/payload release/dist

archive_records="release/archive-records.ndjson"
: > "${archive_records}"

record_archive() {
  local kind="$1"
  local format="$2"
  local file="$3"
  local sha="$4"
  local os="${5:-}"
  local arch="${6:-}"

  @jq@/bin/jq -n \
    --arg kind "${kind}" \
    --arg format "${format}" \
    --arg file "${file}" \
    --arg sha256 "${sha}" \
    --arg os "${os}" \
    --arg arch "${arch}" \
    '{
      kind: $kind,
      format: $format,
      file: $file,
      sha256: $sha256
    } + (if $os == "" then {} else { os: $os } end) + (if $arch == "" then {} else { arch: $arch } end)' \
    >> "${archive_records}"
}

bundle_name="@packageName@_@tag@"
bundle_root="release/payload/${bundle_name}"

mkdir -p "${bundle_root}"

while IFS= read -r entry; do
  [ -n "${entry}" ] || continue
  if [ ! -e "${entry}" ]; then
    printf 'missing required release entry: %s\n' "${entry}" >&2
    exit 1
  fi
  cp -R --parents "${entry}" "${bundle_root}"
done < "@layoutEntriesFile@"

chmod -R u=rwX,go=rX "${bundle_root}"
find "${bundle_root}" -exec touch -h -d '1980-01-01T00:00:00Z' {} +

mapfile -t bundle_formats < "@bundleFormatsFile@"

for archive_ext in "${bundle_formats[@]}"; do
  [ -n "${archive_ext}" ] || continue

  archive_file="${bundle_name}.${archive_ext}"

  case "${archive_ext}" in
    zip)
      (
        cd release/payload
        find "${bundle_name}" -print | sort | @zip@/bin/zip -X -q "../dist/${archive_file}" -@
      )
      ;;
    tar.gz)
      (
        cd release/payload
        @gnutar@/bin/tar --format=gnu --sort=name --mtime='UTC 1980-01-01' --owner=0 --group=0 --numeric-owner -cf - "${bundle_name}" \
          | @gzip@/bin/gzip -n > "../dist/${archive_file}"
      )
      ;;
    *)
      printf 'unsupported archive format: %s\n' "${archive_ext}" >&2
      exit 1
      ;;
  esac

  archive_sha="$(@coreutils@/bin/sha256sum "release/dist/${archive_file}" | cut -d ' ' -f 1)"
  record_archive "repo_bundle" "${archive_ext}" "${archive_file}" "${archive_sha}"
done

mapfile -t binaries < "@binariesFile@"
mapfile -t targets < "@targetsFile@"

for target in "${targets[@]}"; do
  [ -n "${target}" ] || continue
  read -r goos goarch archive_ext <<<"${target}"

  archive_base="@packageName@_@tag@_${goos}_${goarch}"
  package_dir="release/payload/${archive_base}"
  bin_dir="${package_dir}/bin"
  mkdir -p "${bin_dir}"

  for binary in "${binaries[@]}"; do
    [ -n "${binary}" ] || continue
    GOOS="${goos}" GOARCH="${goarch}" go build -ldflags="-s -w" -o "${bin_dir}/${binary}" "./cmd/${binary}"
  done

  cp LICENSE NOTICE README.md "${package_dir}/"
  chmod -R u=rwX,go=rX "${package_dir}"
  find "${package_dir}" -exec touch -h -d '1980-01-01T00:00:00Z' {} +

  archive_file="${archive_base}.${archive_ext}"

  case "${archive_ext}" in
    tar.gz)
      (
        cd release/payload
        @gnutar@/bin/tar --format=gnu --sort=name --mtime='UTC 1980-01-01' --owner=0 --group=0 --numeric-owner -cf - "${archive_base}" \
          | @gzip@/bin/gzip -n > "../dist/${archive_file}"
      )
      ;;
    zip)
      (
        cd release/payload
        find "${archive_base}" -print | sort | @zip@/bin/zip -X -q "../dist/${archive_file}" -@
      )
      ;;
    *)
      printf 'unsupported archive format: %s\n' "${archive_ext}" >&2
      exit 1
      ;;
  esac

  archive_sha="$(@coreutils@/bin/sha256sum "release/dist/${archive_file}" | cut -d ' ' -f 1)"
  record_archive "binary" "${archive_ext}" "${archive_file}" "${archive_sha}" "${goos}" "${goarch}"
done

manifest_path="release/dist/@packageName@_@tag@_release-manifest.json"

@jq@/bin/jq -s \
  --arg package_name "@packageName@" \
  --arg version "@version@" \
  --arg tag "@tag@" \
  '{
    package_name: $package_name,
    version: $version,
    tag: $tag,
    archives: .
  }' "${archive_records}" > "${manifest_path}"

(
  shopt -s nullglob
  cd release/dist
  release_files=( *.tar.gz *.zip *.json )
  if [ "${#release_files[@]}" -eq 0 ]; then
    printf 'expected release assets in release/dist for checksum generation\n' >&2
    exit 1
  fi
  @coreutils@/bin/sha256sum "${release_files[@]}" > SHA256SUMS
)
