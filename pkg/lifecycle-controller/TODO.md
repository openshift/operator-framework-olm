# Future e2e Test Coverage

The unit tests in this package cover behavioral contracts of the builder
functions and reconciliation logic. The items below should be validated as
end-to-end tests running against a real cluster.

## Happy-path test (covers multiple behaviors at once)

A single test can validate most of the controller's output by exercising it
end-to-end:

1. Create a CatalogSource with a catalog image containing lifecycle schema blobs.
2. Wait for the lifecycle-server Deployment to become ready.
3. Query the lifecycle-server API through the Service and assert a correct JSON
   response.

A successful response validates:
- Deployment created with the correct catalog image mounted at the correct
  digest
- TLS via serving-cert annotation and secret
- Service routing to the lifecycle-server pods
- RBAC on nonResourceURL grants access
- FBC path correctness

## Separate targeted tests

- **RBAC denial**: Users without RBAC access to the nonResourceURL are denied.
  Query with an unauthorized ServiceAccount and expect 403.
- **NetworkPolicy**: A NetworkPolicy restricts traffic to only the API port
  inbound and API server + DNS outbound.
- **Node affinity**: The server prefers scheduling on the same node as the
  catalog pod.
- **Pod hardening**: The lifecycle-server pods are hardened: run as non-root,
  read-only root filesystem, drop all capabilities, use seccomp
  RuntimeDefault. (Note: the lifecycle-controller's own hardening is in static
  manifests at `manifests/0000_50_olm_08-lifecycle-controller.deployment.yaml`
  and validated by SCC admission.)
- **Cleanup on deletion**: Resources are cleaned up when a CatalogSource is
  deleted (end-to-end).
- **Cleanup on selector mismatch**: Resources are cleaned up when a
  CatalogSource stops matching selectors (end-to-end).
