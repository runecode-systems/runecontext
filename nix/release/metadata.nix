let
  base = rec {
    packageName = "runecontext";
    version = "0.0.0-dev";

    topLevelFiles = [
      "README.md"
      "LICENSE"
      "NOTICE"
      "DCO"
      "CONTRIBUTING.md"
      "SECURITY.md"
      "CODE_OF_CONDUCT.md"
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
  };
in
base
// {
  tag = "v${base.version}";
}
