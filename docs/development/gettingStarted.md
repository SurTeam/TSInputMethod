# Getting started

**Status:** docs foundation **implemented**; Go daemon and adapters **not yet implemented**.

## Prerequisites (when code lands)

- Go toolchain matching the version pinned in `go.mod` (once added).
- Platform SDK only if you work on an adapter (Xcode / Windows SDK / Fcitx5 headers).

## First steps for agents and humans

1. Read [../index.md](../index.md).
2. Read [../../Agent.md](../../Agent.md).
3. Skim [../architecture/overview.md](../architecture/overview.md) and ADR [0001](../decisions/0001-ipc-protobuf-framed.md).
4. Do not add comments to handwritten source; put design notes in `docs/`.

## Repository layout (emerging)

See the source directory table in [../index.md](../index.md). Empty platform and test trees are kept via README/`.gitkeep` placeholders.
