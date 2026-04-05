# function-bucket-name-normalizer

Crossplane composition function for normalizing bucket-related names before they
are consumed by provider-specific compositions.

This directory is kept as a monorepo subproject. Repository-wide CI,
issue templates, and licensing live at the repo root, not here.

Useful local commands:

```shell
go generate ./...
go test ./...
docker build . --tag=function-bucket-name-normalizer:dev
crossplane xpkg build -f package --embed-runtime-image=function-bucket-name-normalizer:dev
```

Relevant files:

- `fn.go`: function logic
- `fn_test.go`: unit tests
- `input/`: function input type definitions
- `package/crossplane.yaml`: Crossplane function package metadata
- `example/`: local function examples you can use while iterating
