# Database Package

This package exposes a namespaced `XDatabase` abstraction for PostgreSQL workloads.

It follows the same package shape as `bucket`:

- one XRD: `XDatabase`
- two provider-specific compositions:
  - AWS Aurora PostgreSQL
  - Azure Database for PostgreSQL Flexible Server
- example XRs that pin a specific composition with `spec.crossplane.compositionRef`

The abstraction supports the main database knobs in the XRD definition:

- storage sizing
- read replica count
- PostgreSQL version
- admin credentials
- backup retention
- public access
- provider-specific compute sizing

Each composition also publishes a namespaced `ConfigMap` named `<xr-name>-database` with the resolved endpoint information, and the primary managed resource writes connection details to `<xr-name>-db-connection`.

Before applying either example XR, create the admin password secret from `examples/admin-password-secret.yaml`.
