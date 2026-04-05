# function-bucket-name-normalizer

Crossplane composition function for deriving a normalized bucket or container
name from a source field on the composite resource and writing the result back
to a target field on the desired composite resource.

The function currently:

- reads a string from `sourceFieldPath` on the XR
- lowercases the value and replaces unsupported characters with `-`
- collapses repeated separators and trims leading or trailing separators
- enforces `maxLength`
- writes the final name to `targetFieldPath`

This is intended for fields like `status.bucketName`, so later composition
steps can patch the normalized name into provider resources or ConfigMaps.

Example input:

```yaml
apiVersion: functions.devops-autopilot-crossplane-packages.io/v1beta1
kind: NameNormalizerInput
provider: aws
sourceFieldPath: spec.parameters.namePrefix
targetFieldPath: status.bucketName
maxLength: 63
```

For example, `MyBucket_Example` becomes `mybucket-example`.

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
