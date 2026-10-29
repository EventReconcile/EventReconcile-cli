# Local contributor instructions

- Keep user-facing parsing, diagnostics, exit codes and safe file output here.
- Do not copy reconciliation algorithms from `EventReconcile-core`.
- Do not import source-specific adapters; consume the versioned core JSON contract.
- Keep offline operation as the default. Never add automatic network fallback.
- Run `go vet ./...` and `go test -count=1 ./...` before handing off changes.
