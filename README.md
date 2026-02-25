# helm-guard

![CI](https://github.com/r3based/helm-guard/actions/workflows/ci.yml/badge.svg)

> Understand your Helm deployment **before** you apply it.

`helm-guard` is a CLI tool that analyzes rendered Helm manifests and highlights
potential risks, misconfigurations, and operational concerns **before deployment**.

It helps DevOps engineers, SREs, and platform teams answer a simple question:

> “What exactly will this Helm chart deploy — and is it safe?”

---

## Why helm-guard?

Helm charts can be complex:
- Thousands of lines of `values.yaml`
- Multiple overrides
- Nested subcharts
- Hidden defaults

Running `helm install --dry-run` shows raw manifests — but does not:
- Aggregate resource usage
- Highlight Single Points of Failure
- Detect missing probes
- Warn about unsafe container settings
- Summarize networking exposure

`helm-guard` bridges that gap.

---

## Features (Planned & Current Scope)

### ✔ Render-aware analysis
- Uses `helm template`
- Analyzes actual rendered Kubernetes manifests
- Works with any Helm chart

### ✔ Workload summary
- Deployments / StatefulSets / DaemonSets
- Replica counts
- Container images
- Resource requests & limits

### ✔ Networking overview
- Services (ClusterIP / NodePort / LoadBalancer)
- Ingress exposure
- TLS presence

### ✔ Risk & misconfiguration detection
Examples:
- `image: latest`
- Missing resource limits
- Missing readiness probes
- Single replica exposed via Ingress
- Privileged containers
- hostNetwork enabled

### ✔ CI-ready
- Exit with non-zero code on selected severity
- JSON output for pipelines
- Markdown report generation

---

## Installation

### From source

```bash
git clone https://github.com/r3based/helm-guard.git
cd helm-guard
go build -o helm-guard ./cmd/helm-guard
````

Requirements:

* Go 1.21+
* Helm installed and available in `$PATH`

---

## Usage

### Basic analysis

```bash
helm-guard analyze ./chart -f values.yaml
```

### With multiple values files

```bash
helm-guard analyze ./chart -f values.yaml -f prod.yaml
```

### JSON output

```bash
helm-guard analyze ./chart -f values.yaml --output json
```

### Fail CI on high severity findings

```bash
helm-guard analyze ./chart -f values.yaml --fail-on high
```

---

## Example Output

```
Rendered objects: 12

Workloads:
- api (Deployment)
  replicas: 2
  image: myapp:1.2.3
  cpu: 200m / 500m
  memory: 256Mi / 512Mi

Networking:
- Ingress: api.example.com
- Service: ClusterIP

Warnings:
[HIGH] api: Missing readiness probe
[MEDIUM] api: No resource limits defined
```

---

## Architecture

`helm-guard` follows a simple pipeline:

```
helm template
        ↓
Parse Kubernetes manifests
        ↓
Build internal model
        ↓
Apply rules engine
        ↓
Generate report
```

Project structure:

```
cmd/helm-guard       → CLI entrypoint
internal/cli         → Cobra commands
internal/helm        → Helm rendering
internal/kube        → YAML parsing
internal/model       → Internal workload model
internal/rules       → Risk detection engine
internal/report      → Output formatting
```

---

## Design Principles

* **Render-first** (no guessing from values.yaml)
* **No external data sharing**
* **Deterministic analysis (no AI, no heuristics)**
* **CI-friendly**
* **Extensible rule engine**

---

## Contributing

Contributions are welcome.

To run locally:

```bash
go run ./cmd/helm-guard analyze ./chart -f values.yaml
```

Please:

* Keep rules deterministic
* Add tests for new rules
* Follow Go formatting standards (`go fmt`)

---

## License

GNU License

---

## Author

Created by [@r3based](https://github.com/r3based)

If this tool helps you, consider starring the repository.
