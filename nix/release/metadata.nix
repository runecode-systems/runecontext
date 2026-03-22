let
  base = rec {
    packageName = "runecontext";
    version = "0.1.0-alpha.5";

    topLevelFiles = [
      "README.md"
      "LICENSE"
      "NOTICE"
      "DCO"
      "CONTRIBUTING.md"
      "SECURITY.md"
      "CODE_OF_CONDUCT.md"
      "go.mod"
      "go.sum"
      "flake.nix"
      "flake.lock"
      "justfile"
    ];

    topLevelDirectories = [
      "docs"
      "core"
      "adapters"
      "schemas"
      "fixtures"
      "cmd"
      "internal"
      "tools"
      "nix"
    ];

    layoutEntries = topLevelFiles ++ topLevelDirectories;

    bundleFormats = [
      {
        archive = "tar.gz";
      }
      {
        archive = "zip";
      }
    ];

    binaries = [
      "runectx"
    ];

    targets = [
      {
        goos = "linux";
        goarch = "amd64";
        archive = "tar.gz";
      }
      {
        goos = "linux";
        goarch = "arm64";
        archive = "tar.gz";
      }
      {
        goos = "darwin";
        goarch = "amd64";
        archive = "tar.gz";
      }
      {
        goos = "darwin";
        goarch = "arm64";
        archive = "tar.gz";
      }
    ];
  };
in
base
// {
  tag = "v${base.version}";
}
