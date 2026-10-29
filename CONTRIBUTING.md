# Contributing

Run `go vet ./...` and `go test -count=1 ./...` before submitting changes. Keep the CLI thin: parsing, bounded local I/O, diagnostics, safe report writes and exit codes belong here; reconciliation algorithms belong in core.
