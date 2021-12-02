# Cluster Runtime Constraints

Cluster runtime constraints are base constraints that always get applied to the resolution process to avoid installing
packages that might be unsuitable for the cluster (consumes too many resources, wrong kubernetes/ocp version, etc.). 
Currently, there's no first class way to do this. Until we design the canonical way to define
cluster runtime constraints, we are making availing a stopgap solution for IBM and OCP, we:
  1. Are adding a file to the downstream OLM image that includes the runtime constraints
  2. Have modified the upstream catalog operator to load the runtime constraints in to the resolver if a `RUNTIME_CONSTRAINTS` environment variable is defined
  3. Update the deployment manifests for the olm catalog operator deployment to add the `RUNTIME_CONSTRAINTS` environment variable

The upstream PR enabling this behavior is [https://github.com/operator-framework/operator-lifecycle-manager/pull/2498](#2498).
