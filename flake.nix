{
  description = "RuneContext dev environment and canonical release builder";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.11";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    let
      releaseMetadata = import ./nix/release/metadata.nix;
    in
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = import nixpkgs { inherit system; };
        goToolchain = pkgs.go_1_25 or (throw "nixpkgs must provide go_1_25 for RuneContext release builds");

        releaseArtifacts = import ./nix/packages/release-artifacts.nix {
          inherit (pkgs) lib;
          inherit pkgs releaseMetadata self;
        };

        devShell = import ./nix/dev-shell.nix {
          inherit goToolchain pkgs;
        };

        checks = import ./nix/checks.nix {
          inherit (pkgs) lib;
          inherit
            devShell
            pkgs
            releaseArtifacts
            releaseMetadata
            self
            system
            ;
        };
      in
      {
        formatter = pkgs.nixfmt-rfc-style;

        devShells.default = devShell;

        packages = {
          default = releaseArtifacts;
          release-artifacts = releaseArtifacts;
        };

        inherit checks;
      }
    )
    // {
      lib.release = releaseMetadata;
    };
}
