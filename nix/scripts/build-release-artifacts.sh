#!/usr/bin/env bash
set -euo pipefail

umask 022

export SOURCE_DATE_EPOCH=315532800
export TZ=UTC
export LC_ALL=C

mkdir -p release/payload release/dist

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

  case "${archive_ext}" in
    zip)
      (
        cd release/payload
        find "${bundle_name}" -print | sort | @zip@/bin/zip -X -q "../dist/${bundle_name}.zip" -@
      )
      ;;
    tar.gz)
      (
        cd release/payload
        @gnutar@/bin/tar --format=gnu --sort=name --mtime='UTC 1980-01-01' --owner=0 --group=0 --numeric-owner -cf - "${bundle_name}" \
          | @gzip@/bin/gzip -n > "../dist/${bundle_name}.tar.gz"
      )
      ;;
    *)
      printf 'unsupported archive format: %s\n' "${archive_ext}" >&2
      exit 1
      ;;
  esac
done

manifest_path="release/dist/@packageName@_@tag@_release-manifest.json"

tar_sha=""
zip_sha=""
if [ -f "release/dist/${bundle_name}.tar.gz" ]; then
  tar_sha="$(@coreutils@/bin/sha256sum "release/dist/${bundle_name}.tar.gz" | cut -d ' ' -f 1)"
fi
if [ -f "release/dist/${bundle_name}.zip" ]; then
  zip_sha="$(@coreutils@/bin/sha256sum "release/dist/${bundle_name}.zip" | cut -d ' ' -f 1)"
fi

@jq@/bin/jq -n \
  --arg package_name "@packageName@" \
  --arg version "@version@" \
  --arg tag "@tag@" \
  --arg tar_file "${bundle_name}.tar.gz" \
  --arg tar_sha "${tar_sha}" \
  --arg zip_file "${bundle_name}.zip" \
  --arg zip_sha "${zip_sha}" \
  '{
    package_name: $package_name,
    version: $version,
    tag: $tag,
    archives: [
      { format: "tar.gz", file: $tar_file, sha256: $tar_sha },
      { format: "zip", file: $zip_file, sha256: $zip_sha }
    ] | map(select(.sha256 != ""))
  }' > "${manifest_path}"

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
