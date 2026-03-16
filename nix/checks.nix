{
  devShell,
  lib,
  pkgs,
  releaseArtifacts,
  releaseMetadata,
  self,
  system,
}:

let
  declaredLayoutEntries = pkgs.writeText "runecontext-layout-check.txt" (
    lib.concatStringsSep "\n" releaseMetadata.layoutEntries + "\n"
  );
in
{
  dev-shell = devShell;

  nix-format =
    pkgs.runCommand "nix-format-check"
      {
        nativeBuildInputs = [
          pkgs.fd
          pkgs.nixfmt-rfc-style
        ];
      }
      ''
        files=("${self}/flake.nix")
        while IFS= read -r file; do
          files+=("$file")
        done < <(${pkgs.fd}/bin/fd --extension nix --type f . ${self}/nix)

        nixfmt --check "''${files[@]}"
        touch "$out"
      '';

  layout =
    pkgs.runCommand "layout-check"
      {
        nativeBuildInputs = [
          pkgs.coreutils
          pkgs.diffutils
        ];
      }
      ''
        actual="$TMPDIR/actual-layout"
        declared="$TMPDIR/declared-layout"

        while IFS= read -r entry; do
          [ -n "$entry" ] || continue
          if [ ! -e "${self}/$entry" ]; then
            printf 'missing required layout entry: %s\n' "$entry" >&2
            exit 1
          fi
          printf '%s\n' "$entry"
        done < ${declaredLayoutEntries} | sort > "$actual"

        sort ${declaredLayoutEntries} > "$declared"

        diff -u "$declared" "$actual"
        touch "$out"
      '';
}
// lib.optionalAttrs (system == "x86_64-linux") {
  # Release artifacts are checked on the canonical Linux release architecture.
  release-artifacts = releaseArtifacts;
}
