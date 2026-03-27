{
  goToolchain,
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

  binariesFile = pkgs.writeText "runecontext-release-binaries.txt" (
    lib.concatStringsSep "\n" releaseMetadata.binaries + "\n"
  );

  targetsFile = pkgs.writeText "runecontext-release-targets.txt" (
    lib.concatMapStringsSep "\n" (
      target: "${target.goos} ${target.goarch} ${target.archive}"
    ) releaseMetadata.targets
    + "\n"
  );

  schemaBundlesFile = pkgs.writeText "runecontext-release-schema-bundles.json" (
    builtins.toJSON releaseMetadata.schemaBundles + "\n"
  );

  adapterPacksFile = pkgs.writeText "runecontext-release-adapter-packs.json" (
    builtins.toJSON releaseMetadata.adapterPacks + "\n"
  );

  installerScriptsFile = pkgs.writeText "runecontext-release-installer-scripts.txt" (
    lib.concatStringsSep "\n" releaseMetadata.installerScripts + "\n"
  );

  buildScript = pkgs.writeText "build-release-artifacts.sh" (
    renderTemplate ../scripts/build-release-artifacts.sh {
      packageName = releaseMetadata.packageName;
      tag = releaseMetadata.tag;
      version = releaseMetadata.version;
      layoutEntriesFile = layoutEntriesFile;
      bundleFormatsFile = bundleFormatsFile;
      schemaBundlesFile = schemaBundlesFile;
      adapterPacksFile = adapterPacksFile;
      installerScriptsFile = installerScriptsFile;
      binariesFile = binariesFile;
      targetsFile = targetsFile;
      coreutils = pkgs.coreutils;
      findutils = pkgs.findutils;
      jq = pkgs.jq;
      gnutar = pkgs.gnutar;
      gzip = pkgs.gzip;
      zip = pkgs.zip;
    }
  );
in
pkgs.buildGoModule {
  pname = "${releaseMetadata.packageName}-release-artifacts";
  version = releaseMetadata.version;
  src = releaseSource;
  go = goToolchain;
  vendorHash = "sha256-Eepto6l5Z5KQGcKxhUNao2Uy5u6Gle3clqh43M+hbrs=";
  doCheck = false;
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
