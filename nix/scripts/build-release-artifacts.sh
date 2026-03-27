#!/usr/bin/env bash
set -euo pipefail

umask 022

export CGO_ENABLED=0
export GOFLAGS="-trimpath -mod=vendor"
export SOURCE_DATE_EPOCH=315532800
export TZ=UTC
export LC_ALL=C

coreutils='@coreutils@/bin'
findutils='@findutils@/bin'

"${coreutils}/mkdir" -p release/payload release/dist

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

"${coreutils}/mkdir" -p "${bundle_root}"

while IFS= read -r entry; do
  [ -n "${entry}" ] || continue
  if [ ! -e "${entry}" ]; then
    printf 'missing required release entry: %s\n' "${entry}" >&2
    exit 1
  fi
  "${coreutils}/cp" -R --parents "${entry}" "${bundle_root}"
done < "@layoutEntriesFile@"

"${coreutils}/chmod" -R u=rwX,go=rX "${bundle_root}"
"${findutils}/find" "${bundle_root}" -exec "${coreutils}/touch" -h -d '1980-01-01T00:00:00Z' {} +

mapfile -t bundle_formats < "@bundleFormatsFile@"

for archive_ext in "${bundle_formats[@]}"; do
  [ -n "${archive_ext}" ] || continue

  archive_file="${bundle_name}.${archive_ext}"

  case "${archive_ext}" in
    zip)
      (
        cd release/payload
        "${findutils}/find" "${bundle_name}" -print | "${coreutils}/sort" | @zip@/bin/zip -X -q "../dist/${archive_file}" -@
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

  archive_sha="$(@coreutils@/bin/sha256sum "release/dist/${archive_file}" | "${coreutils}/cut" -d ' ' -f 1)"
  record_archive "repo_bundle" "${archive_ext}" "${archive_file}" "${archive_sha}"
done

process_pack_archives() {
  local kind="$1"
  local entries_json="$2"

  if [ ! -s "${entries_json}" ]; then
    return
  fi

  if ! @jq@/bin/jq -e '
    type == "array"
    and all(
      .[];
      type == "object"
      and (.name | type == "string")
      and (.entries | type == "array")
      and all(.entries[]; type == "string" and length > 0)
    )
  ' "${entries_json}" >/dev/null; then
    printf 'invalid pack metadata: %s\n' "${entries_json}" >&2
    exit 1
  fi

  local -a packs
  if ! mapfile -t packs < <(@jq@/bin/jq -ce '.[]' "${entries_json}"); then
    printf 'failed to parse pack metadata: %s\n' "${entries_json}" >&2
    exit 1
  fi

  for pack in "${packs[@]}"; do
    [ -n "${pack}" ] || continue
    local name
    if ! name="$(@jq@/bin/jq -er '.name' <<<"${pack}")"; then
      printf 'invalid pack metadata (missing name): %s\n' "${entries_json}" >&2
      exit 1
    fi
    if [[ ! "${name}" =~ ^[a-z0-9][a-z0-9-]*$ ]]; then
      printf 'invalid pack name: %s\n' "${name}" >&2
      exit 1
    fi
    local -a entries
    if ! mapfile -t entries < <(@jq@/bin/jq -er '.entries[]' <<<"${pack}"); then
      printf 'invalid pack metadata (entries): %s\n' "${entries_json}" >&2
      exit 1
    fi

    local pack_root
    pack_root="release/payload/${name}"
    "${coreutils}/rm" -rf "${pack_root}"
    "${coreutils}/mkdir" -p "${pack_root}"

    for entry in "${entries[@]}"; do
      [ -n "${entry}" ] || continue
      if [ ! -e "${entry}" ]; then
        printf 'missing required release entry: %s\n' "${entry}" >&2
        exit 1
      fi
      "${coreutils}/cp" -R --parents "${entry}" "${pack_root}"
    done

    "${coreutils}/chmod" -R u=rwX,go=rX "${pack_root}"
    "${findutils}/find" "${pack_root}" -exec "${coreutils}/touch" -h -d '1980-01-01T00:00:00Z' {} +

    (cd release/payload && @gnutar@/bin/tar --format=gnu --sort=name --mtime='UTC 1980-01-01' --owner=0 --group=0 --numeric-owner -cf - "${name}" | @gzip@/bin/gzip -n > "../dist/${name}.tar.gz")

    local pack_sha
    pack_sha="$(@coreutils@/bin/sha256sum "release/dist/${name}.tar.gz" | "${coreutils}/cut" -d ' ' -f 1)"
    record_archive "${kind}" "tar.gz" "${name}.tar.gz" "${pack_sha}"
  done
}

process_pack_archives "schema_bundle" "@schemaBundlesFile@"
process_pack_archives "adapter_pack" "@adapterPacksFile@"

mapfile -t binaries < "@binariesFile@"
mapfile -t targets < "@targetsFile@"

for target in "${targets[@]}"; do
  [ -n "${target}" ] || continue
  read -r goos goarch archive_ext <<<"${target}"

  archive_base="@packageName@_@tag@_${goos}_${goarch}"
  package_dir="release/payload/${archive_base}"
  bin_dir="${package_dir}/bin"
  "${coreutils}/mkdir" -p "${bin_dir}"

    for binary in "${binaries[@]}"; do
      [ -n "${binary}" ] || continue
      ldflags_version="@tag@"
      ldflags_version="${ldflags_version#v}"
      GOOS="${goos}" GOARCH="${goarch}" go build -ldflags="-s -w -X github.com/runecode-systems/runecontext/internal/cli.runecontextVersion=${ldflags_version}" -o "${bin_dir}/${binary}" "./cmd/${binary}"
    done

  "${coreutils}/cp" LICENSE NOTICE README.md "${package_dir}/"
  "${coreutils}/chmod" -R u=rwX,go=rX "${package_dir}"
  "${findutils}/find" "${package_dir}" -exec "${coreutils}/touch" -h -d '1980-01-01T00:00:00Z' {} +

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
        "${findutils}/find" "${archive_base}" -print | "${coreutils}/sort" | @zip@/bin/zip -X -q "../dist/${archive_file}" -@
      )
      ;;
    *)
      printf 'unsupported archive format: %s\n' "${archive_ext}" >&2
      exit 1
      ;;
  esac

  archive_sha="$(@coreutils@/bin/sha256sum "release/dist/${archive_file}" | "${coreutils}/cut" -d ' ' -f 1)"
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
