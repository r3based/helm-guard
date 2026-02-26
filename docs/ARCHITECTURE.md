# Architecture and tech debt

## Current layout

```
cmd/helm-guard          → entrypoint
internal/
  cli/                  → Cobra: analyze, rules; global flags
  helm/                 → helm template (exec)
  kube/                 → ParseManifests (YAML → []unstructured)
  model/                → Build(model), types, PublicServices/WorkloadsForService
  profile/              → Profile, LoadFromYAML, ResolveProfile, BuildEffectiveProfile, ToEngineOptions
  report/               → Pretty, (json/markdown stubs)
  rules/                → Rule, Engine, Finding, Severity
  rules/builtin/        → All(), concrete rules
tests/
  model/                → tests for model (build, exposure)
  profile/              → tests for profile
  builtin/              → tests for builtin rules
```

Pipeline: `helm template → ParseManifests → model.Build → profile resolve → engine.Run → report.Pretty`.

---

## Weak spots

### 1. CLI does too much

`analyze.RunE` owns the whole flow: helm, parse, model, profile load/resolve, engine, report, exit code. Hard to unit-test the pipeline or reuse it from another entrypoint (e.g. library or another CLI).

**Improvement:** Extract an `analyzer` or `pipeline` package that takes config (chart path, values, profile name/file, disable, failOn) and returns (model, findings, effective profile). CLI only parses flags and calls this API. Then you can test the pipeline with in-memory inputs and add e2e tests.

### 2. Monolithic model build

`model.Build` is one big switch and many `build*` helpers in one file. Adding a new resource type touches the same file and the same switch; the file is large and will keep growing.

**Improvement:** Introduce a registry of kind → builder (e.g. `RegisterKind("Deployment", buildWorkload)`). Each builder in its own file (`build_workload.go`, `build_service.go`, …). Optional: builder returns a “fragment” (e.g. which slice to append to) so the core loop stays small and new types don’t require changing the central switch.

### 3. Report tied to full model

`Pretty` knows about every model slice (Workloads, Services, Ingresses, ConfigMaps, …) and prints them in a fixed order. Any new field or type forces editing `report/pretty.go`. JSON/markdown are stubs.

**Improvement:** Define a small “report view” (e.g. struct or interface) that the report layer consumes, and build that view from the model in one place. Then add `--output json` with a stable schema and, if needed, a separate markdown formatter, both using the same view so the model shape doesn’t leak into every formatter.

### 4. No abstraction for “render” and “parse”

Helm rendering and kube parsing are used only via `helm.Template` and `kube.ParseManifests`. Tests that need “rendered” manifests construct `[]unstructured` by hand, which is good; but there is no interface, so swapping (e.g. kustomize or pre-rendered YAML) would require call-site changes.

**Improvement (optional):** Introduce a small `Renderer` interface (e.g. `Render(ctx) ([]byte, error)`) and/or `ManifestParser` (`Parse([]byte) ([]unstructured, error)`). Default implementations call helm and kube. This lowers tech debt if you later need multiple backends; if you stay single-backend, the gain is small.

### 5. Global CLI flags

Flags are package-level vars in `cli`. Testing a command in isolation or running two analyzes in one process is awkward.

**Improvement:** Pass a config struct into the command’s RunE (or into the pipeline), built from flags in `RunE`. No global state for flags; easier to test and to reuse from tests or another CLI.

### 6. Profile file open in CLI

Profile file is opened in `analyze.go` with `os.Open(profileFile)`. Fine for CLI, but the pipeline is then tied to the filesystem.

**Improvement:** Pipeline accepts an `io.Reader` for the profile (or an already-loaded `*profile.Profile`). CLI opens the file and passes the reader (or the parsed profile); tests pass a buffer or a pre-built profile.

### 7. Engine and rules depend on full model

Rules receive `model.Model` and use whatever they need. That’s flexible but couples every rule to the full model. Not a problem at current scale, but if the model grows a lot, “views” (e.g. only workloads + services + ingresses for exposure rules) could reduce coupling.

**Improvement (later):** Only if the model becomes very large or rules need a different “slice” of data: define narrow interfaces or view structs that the engine passes to specific rules. Low priority until the model is much bigger.

### 8. Duplication between report and possible JSON

If you add a full JSON report, you’ll duplicate the list of “what to include” (workloads, services, …) and their field names. That increases the chance of drift and bugs.

**Improvement:** Single “report model” or DTO (e.g. `ReportInput` or view struct) produced from `model.Model` + `findings` + metadata (profile name, failOn). Pretty and JSON both consume that; adding a field happens in one place.

---

## Suggested order of work (to reduce tech debt)

1. **Extract pipeline**  
   New package (e.g. `internal/analyzer` or `internal/run`): function that takes chart path + values + profile options + disable/failOn, returns model + findings + effective profile (and optionally rendered count). CLI only parses flags and calls it. Add one integration test that runs the pipeline on a tiny chart in `testdata/`.

2. **Pass config instead of globals**  
   Refactor CLI to build a config struct from flags and pass it into the pipeline and into report/exit-code logic. Removes global flag vars and makes testing and reuse easier.

3. **Profile from reader**  
   Pipeline accepts profile content via `io.Reader` (or pre-parsed `*profile.Profile`). CLI opens `--profile-file` and passes the reader. Keeps I/O at the edge and keeps the pipeline testable without touching the filesystem.

4. **Stable JSON output**  
   Define a small, versioned JSON schema (e.g. `schemaVersion`, `summary`, `findings`, `profile`). Implement output in `report` using the same data that Pretty uses (or a shared view). Document the schema so CI can rely on it.

5. **Split model build (optional)**  
   If `build.go` becomes hard to maintain, add a builder registry and move each resource type into its own file. Reduces merge conflicts and keeps the “add a new kind” change local.

6. **Report view (optional)**  
   If you add more output formats, introduce a single “report view” built from model + findings and feed it to Pretty, JSON, and any future formatter. Reduces duplication and keeps behavior consistent.

---

## Tests

Tests are under `tests/` in packages `model_test`, `profile_test`, `builtin_test`, so they run against the public API of each package. To run only those:

```bash
go test ./tests/...
```

To run everything (including any future tests inside `internal/`):

```bash
go test ./...
```

Adding integration tests that run the full pipeline (e.g. `tests/integration` with a small chart in `testdata/`) will help catch regressions when changing the pipeline or the model.
