{
  lib,
  pkgs,
  releaseMetadata,
  self,
}:

let
  renderTemplate =
    template: replacements:
    lib.foldlAttrs (
      rendered: name: value:
      lib.replaceStrings [ "@${name}@" ] [ (toString value) ] rendered
    ) (builtins.readFile template) replacements;

  releaseSource = lib.cleanSourceWith {
    src = self;
    filter =
      path: _type:
      let
        root = toString self;
        pathString = toString path;
        relativePath = if pathString == root then "." else lib.removePrefix "${root}/" pathString;
        keepPrefixes = releaseMetadata.topLevelDirectories;
        matchesPrefix =
          prefix:
          relativePath == prefix
          || lib.hasPrefix "${prefix}/" relativePath
          || lib.hasPrefix "${relativePath}/" prefix;
      in
      relativePath == "."
      || lib.elem relativePath releaseMetadata.topLevelFiles
      || lib.any matchesPrefix keepPrefixes;
  };

  layoutEntriesFile = pkgs.writeText "runecontext-release-layout.txt" (
    lib.concatStringsSep "\n" releaseMetadata.layoutEntries + "\n"
  );

  bundleFormatsFile = pkgs.writeText "runecontext-release-bundle-formats.txt" (
    lib.concatMapStringsSep "\n" (bundle: bundle.archive) releaseMetadata.bundleFormats + "\n"
  );

  buildScript = pkgs.writeText "build-release-artifacts.sh" (
    renderTemplate ../scripts/build-release-artifacts.sh {
      packageName = releaseMetadata.packageName;
      tag = releaseMetadata.tag;
      version = releaseMetadata.version;
      layoutEntriesFile = layoutEntriesFile;
      bundleFormatsFile = bundleFormatsFile;
      coreutils = pkgs.coreutils;
      jq = pkgs.jq;
      gnutar = pkgs.gnutar;
      gzip = pkgs.gzip;
      zip = pkgs.zip;
    }
  );
in
pkgs.stdenvNoCC.mkDerivation {
  pname = "${releaseMetadata.packageName}-release-artifacts";
  version = releaseMetadata.version;
  src = releaseSource;
  strictDeps = true;

  nativeBuildInputs = [
    pkgs.bash
    pkgs.coreutils
    pkgs.findutils
    pkgs.jq
    pkgs.gnutar
    pkgs.gzip
    pkgs.zip
  ];

  buildPhase = ''
    runHook preBuild
    bash ${buildScript}
    runHook postBuild
  '';

  installPhase = ''
    runHook preInstall

    mkdir -p "$out"
    cp -R release/dist "$out/dist"
    cp -R release/payload "$out/payload"

    runHook postInstall
  '';
}
