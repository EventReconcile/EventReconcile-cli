# EventReconcile-cli

Go user-facing CLI consuming `EventReconcile-core` and owning local input/output, diagnostics and exit codes.

Usage:

```text
eventreconcile --input dataset.json --output report.json
```

Exit codes: `0` PASS, `1` FAIL, `2` INCONCLUSIVE, `3` usage/input/output error. Reports are written with create-only semantics and existing files are never overwritten.

Offline development:

```powershell
go vet ./...
go test -count=1 ./...
go run ./cmd/eventreconcile --input input.json --output report.json
```

The CLI consumes a core-compatible JSON document. It does not import `EventReconcile-adapters`; use an adapter normalizer as a separate step.
