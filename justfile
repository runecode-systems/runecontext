default:
  @just --list

fmt:
  nixfmt flake.nix $(fd --extension nix --type f . nix)

lint:
  just layout-check

test:
  go test ./...

release-check:
  nix build --no-link .#release-artifacts

layout-check:
  test -f README.md
  test -f flake.nix
  test -f flake.lock
  test -f justfile
  test -d core
  test -d adapters
  test -d docs
  test -d schemas
  test -d fixtures
  test -d cmd/runectx
  test -d internal
  test -d tools/releasebuilder

check:
  nix flake check --no-write-lock-file

ci: lint test

nix-ci: lint test release-check check

release:
  nix build --no-link .#release-artifacts

dev:
  @just --list
