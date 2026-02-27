# Resolution Trace Data Collection Guide

All instrumented log lines use the prefix `RESOLUTION_TRACE` with key=value pairs.
Every line includes `phase=` and `event=BEGIN|DONE`. DONE lines include `duration_ms=`.

## Collecting Logs

```bash
# From an OLM pod on an OpenShift cluster
oc logs deploy/olm-operator -n openshift-operator-lifecycle-manager > olm.log

# Extract only trace lines
grep RESOLUTION_TRACE olm.log > traces.log
```

## Phase Reference

### Orchestration (step_resolver.go)

| Phase | Type | Key Fields (DONE) |
|---|---|---|
| `resolve_steps` | wall clock | `namespace`, `subscription_count`, `error` (on failure) |
| `list_subscriptions` | I/O (API) | `namespace`, `subscription_count` |
| `list_operator_groups` | I/O (API) | `namespace`, `count` |
| `core_resolve` | hybrid | `namespace`, `subscription_count`, `namespace_count`, `operator_count` |
| `generate_steps` | compute | `namespace`, `operator_count`, `step_count`, `bundle_lookup_count` |

### Core Resolution (resolver.go)

| Phase | Type | Key Fields (DONE) |
|---|---|---|
| `cache_namespaced` | I/O coord | `namespace`, `namespace_count` |
| `existing_variables` | compute | `namespace`, `entry_count`, `variable_count` |
| `subscription_variables` | compute | `namespace`, `subscription_count`, `total_variables` |
| `single_sub_variables` | compute | `namespace`, `subscription`, `package`, `channel`, `catalog`, `has_current`, `variable_count` |
| `add_invariants` | compute | `namespace`, `variable_count_before`, `variable_count_after`, `invariants_added` |
| `sat_solve` | compute | `namespace`, `variable_count`, `subscription_count`, `solved_count` |
| `extract_results` | compute | `namespace`, `operator_count` |

### Per-Subscription Detail (resolver.go)

| Phase | Type | Key Fields (DONE) |
|---|---|---|
| `cache_find_entries` | compute | `catalog`, `entry_count` |
| `sort_channel` | compute | `entry_count`, `sorted_count` |
| `build_candidates` | compute | `sorted_bundle_count`, `candidate_count` |
| `bundle_variables_dfs` | compute | `initial_stack_size`, `visited_count`, `variables_count`, `bundles_processed`, `dependency_predicate_count`, `total_dep_bundles` |

### Cache (cache.go)

| Phase | Type | Key Fields (DONE) |
|---|---|---|
| `cache_populate` | I/O coord | `source_count`, `cache_hit_count`, `cache_miss_count` |

**Note:** On cache misses, `cache_populate` DONE fires asynchronously after all snapshot
goroutines complete, so its `duration_ms` reflects the actual I/O wall time.

### Catalog Snapshots (source_registry.go, source_csvs.go)

| Phase | Type | Key Fields (DONE) |
|---|---|---|
| `catalog_snapshot` | I/O (gRPC) | `catalog`, `bundle_count`, `package_count` |
| `list_bundles` | I/O (gRPC) | `catalog`, `bundle_count`, `package_count`, `get_package_calls`, `get_package_ms` |
| `csv_snapshot` | I/O (API) | `namespace`, `csv_count`, `subscription_count`, `entry_count` |

### SAT Solver (solver/solve.go)

| Phase | Type | Key Fields (DONE) |
|---|---|---|
| `sat_add_constraints` | compute | `variable_count`, `constraint_count`, `literal_count` |
| `sat_search` | compute | `anchor_count`, `assumption_count`, `final_assumption_count`, `outcome` |
| `sat_minimize` | compute | `extras_count`, `excluded_count`, `cardinality_bound`, `result_count` |

---

## Correlation: Linking Phases to the Same Resolution Cycle

The `resolve_steps` phase (in step_resolver.go) injects a `resolution_id=` field
into the logrus logger. This enriched logger is propagated via `Resolver.WithLog()`
to all inner phases, so **all phases** emitted from `step_resolver.go`, `resolver.go`,
and `solver/solve.go` carry `resolution_id` as a logrus field.

Phases in the **cache** and **catalog snapshot** layers (`cache_populate`,
`catalog_snapshot`, `list_bundles`, `csv_snapshot`) use `logrus.StdLogger` (no field
support), so they do not carry `resolution_id`. Correlate these by `namespace=` or
`catalog=` and timestamp proximity.

### Get all phases for a specific resolution cycle by resolution_id

```bash
ID="a1b2c3d4"
grep "resolution_id=$ID" olm.log | grep RESOLUTION_TRACE
```

### Correlate cache/snapshot phases by timestamp window

```bash
NS="openshift-operators"
# 1. Find the resolve_steps BEGIN/DONE timestamps for a namespace
grep "phase=resolve_steps" traces.log | grep "namespace=$NS"

# 2. Use awk to filter traces within that time window
# (all traces between a BEGIN and its matching DONE belong to one cycle)
```

---

## Experiment 1: Resolution Latency vs Subscription Count

**Goal:** How does total resolution time scale with the number of Subscriptions in a namespace?

### Collect total resolution time per namespace

```bash
grep "phase=resolve_steps event=DONE" traces.log \
  | grep -oP 'namespace=\S+ subscription_count=\d+ .*?duration_ms=\d+' \
  | sort
```

### Collect core_resolve time (excludes API listing overhead)

```bash
grep "phase=core_resolve event=DONE" traces.log \
  | grep -oP 'namespace=\S+ subscription_count=\d+ .*?duration_ms=\d+'
```

### Per-subscription breakdown within a namespace

```bash
NS="my-namespace"
grep "phase=single_sub_variables event=DONE" traces.log \
  | grep "namespace=$NS" \
  | grep -oP 'subscription=\S+ package=\S+ catalog=\S+ has_current=\S+ variable_count=\d+ .*?duration_ms=\d+'
```

### Subscription count vs SAT solve time

```bash
grep "phase=sat_solve event=DONE" traces.log \
  | grep -oP 'namespace=\S+ variable_count=\d+ subscription_count=\d+ .*?duration_ms=\d+'
```

---

## Experiment 2: Resolution Latency vs Catalog Size

**Goal:** How does resolution time change based on which catalog is used and how many bundles it contains?

### Catalog snapshot fetch times (gRPC round-trip cost)

```bash
grep "phase=catalog_snapshot event=DONE" traces.log \
  | grep -oP 'catalog=\S+ bundle_count=\d+ package_count=\d+ .*?duration_ms=\d+'
```

### gRPC ListBundles time per catalog (includes GetPackage aggregate)

```bash
grep "phase=list_bundles event=DONE" traces.log \
  | grep -oP 'catalog=\S+ bundle_count=\d+ package_count=\d+ get_package_calls=\d+ get_package_ms=\d+ .*?duration_ms=\d+'
```

### Cache hit/miss rate (are snapshots being reused?)

```bash
grep "phase=cache_populate event=DONE" traces.log \
  | grep -oP 'source_count=\d+ cache_hit_count=\d+ cache_miss_count=\d+ .*?duration_ms=\d+'
```

### Per-subscription: which catalog and how many entries matched

```bash
grep "phase=cache_find_entries event=DONE" traces.log \
  | grep -oP 'catalog=\S+ entry_count=\d+ .*?duration_ms=\d+'
```

---

## Experiment 3: Dependency Graph Complexity

**Goal:** How does the DFS dependency resolution scale with bundles and dependencies?

### DFS work per resolution

```bash
grep "phase=bundle_variables_dfs event=DONE" traces.log \
  | grep -oP 'bundles_processed=\d+ dependency_predicate_count=\d+ total_dep_bundles=\d+ visited_count=\d+ variables_count=\d+ .*?duration_ms=\d+'
```

### Invariant growth (GVK/package conflict constraints)

```bash
grep "phase=add_invariants event=DONE" traces.log \
  | grep -oP 'namespace=\S+ variable_count_before=\d+ variable_count_after=\d+ invariants_added=\d+ .*?duration_ms=\d+'
```

---

## Experiment 4: SAT Solver Internals

**Goal:** How does the SAT problem size affect solve time?

### Problem size (variables, constraints, literals)

```bash
grep "phase=sat_add_constraints event=DONE" traces.log \
  | grep -oP 'variable_count=\d+ constraint_count=\d+ literal_count=\d+ .*?duration_ms=\d+'
```

### Search complexity (anchor count, backtracking)

```bash
grep "phase=sat_search event=DONE" traces.log \
  | grep -oP 'outcome=\d+ anchor_count=\d+ final_assumption_count=\d+ .*?duration_ms=\d+'
```

### Cardinality minimization cost

```bash
grep "phase=sat_minimize event=DONE" traces.log \
  | grep -oP 'extras_count=\d+ excluded_count=\d+ cardinality_bound=\d+ result_count=\d+ .*?duration_ms=\d+'
```

---

## Experiment 5: Installed Operators (Virtual Catalog)

**Goal:** How does the number of already-installed CSVs affect resolution?

### CSV snapshot size and cost

```bash
grep "phase=csv_snapshot event=DONE" traces.log \
  | grep -oP 'namespace=\S+ csv_count=\d+ subscription_count=\d+ entry_count=\d+ .*?duration_ms=\d+'
```

### Existing variables processing time

```bash
grep "phase=existing_variables event=DONE" traces.log \
  | grep -oP 'namespace=\S+ entry_count=\d+ variable_count=\d+ .*?duration_ms=\d+'
```

---

## Filtering by Namespace

Every query above can be scoped to a single namespace:

```bash
NS="openshift-operators"
grep "RESOLUTION_TRACE" traces.log | grep "namespace=$NS"
```

For phases that use `catalog=` instead of `namespace=` (snapshot phases), the catalog
field is formatted as `namespace/name`, so filter with:

```bash
grep "catalog=$NS/" traces.log
```

---

## Quick One-Liner: Full Pipeline Summary for a Namespace

```bash
NS="openshift-operators"
echo "=== Resolution Timeline ==="
grep "RESOLUTION_TRACE.*namespace=$NS.*event=DONE" traces.log \
  | grep -oP 'phase=\S+ .*?duration_ms=\d+' \
  | column -t

echo ""
echo "=== Catalog Snapshots ==="
grep "phase=catalog_snapshot event=DONE" traces.log \
  | grep -oP 'catalog=\S+ bundle_count=\d+ package_count=\d+ .*?duration_ms=\d+' \
  | column -t

echo ""
echo "=== SAT Solver Detail ==="
grep "phase=sat_" traces.log \
  | grep "event=DONE" \
  | grep -oP 'phase=\S+ .*' \
  | column -t
```

---

## Generating a CSV for Analysis

```bash
# Example: subscription count vs resolution time
echo "namespace,subscription_count,duration_ms" > resolution_by_subs.csv
grep "phase=resolve_steps event=DONE" traces.log \
  | sed -E 's/.*namespace=(\S+).*subscription_count=([0-9]+).*duration_ms=([0-9]+).*/\1,\2,\3/' \
  >> resolution_by_subs.csv

# Example: catalog snapshot size vs fetch time
echo "catalog,bundle_count,package_count,duration_ms" > catalog_snapshots.csv
grep "phase=catalog_snapshot event=DONE" traces.log \
  | sed -E 's/.*catalog=(\S+).*bundle_count=([0-9]+).*package_count=([0-9]+).*duration_ms=([0-9]+).*/\1,\2,\3,\4/' \
  >> catalog_snapshots.csv

# Example: SAT problem size vs solve time
echo "namespace,variable_count,subscription_count,solved_count,duration_ms" > sat_solve.csv
grep "phase=sat_solve event=DONE" traces.log \
  | grep -v "error=true" \
  | sed -E 's/.*namespace=(\S+).*variable_count=([0-9]+).*subscription_count=([0-9]+).*solved_count=([0-9]+).*duration_ms=([0-9]+).*/\1,\2,\3,\4,\5/' \
  >> sat_solve.csv

# Example: per-catalog gRPC cost breakdown
echo "catalog,bundle_count,package_count,get_package_calls,get_package_ms,duration_ms" > list_bundles.csv
grep "phase=list_bundles event=DONE" traces.log \
  | sed -E 's/.*catalog=(\S+).*bundle_count=([0-9]+).*package_count=([0-9]+).*get_package_calls=([0-9]+).*get_package_ms=([0-9]+).*duration_ms=([0-9]+).*/\1,\2,\3,\4,\5,\6/' \
  >> list_bundles.csv
```
