package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/EventReconcile/EventReconcile-core/reconcile"
)

const maxInputBytes int64 = 16 * 1024 * 1024

const (
	exitPass = iota
	exitFail
	exitInconclusive
	exitUsage
)

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("eventreconcile", flag.ContinueOnError)
	flags.SetOutput(stderr)
	inputPath := flags.String("input", "", "JSON input dataset")
	outputPath := flags.String("output", "", "report destination; existing files are never overwritten")
	if err := flags.Parse(args); err != nil {
		return exitUsage
	}
	if *inputPath == "" || *outputPath == "" || flags.NArg() != 0 {
		fmt.Fprintln(stderr, "usage: eventreconcile --input dataset.json --output report.json")
		return exitUsage
	}
	if filepath.Clean(*inputPath) == filepath.Clean(*outputPath) {
		fmt.Fprintln(stderr, "input and output paths must differ")
		return exitUsage
	}
	data, err := readBounded(*inputPath)
	if err != nil {
		fmt.Fprintf(stderr, "INPUT_ERROR: %v\n", err)
		return exitUsage
	}
	var input reconcile.Input
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&input); err != nil {
		fmt.Fprintf(stderr, "INPUT_ERROR: invalid JSON: %v\n", err)
		return exitUsage
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		fmt.Fprintln(stderr, "INPUT_ERROR: input must contain exactly one JSON document")
		return exitUsage
	}
	report, err := reconcile.Reconcile(input)
	if err != nil {
		fmt.Fprintf(stderr, "INPUT_ERROR: %v\n", err)
		return exitUsage
	}
	reportBytes, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Fprintf(stderr, "REPORT_ERROR: %v\n", err)
		return exitUsage
	}
	if err := writeNew(*outputPath, append(reportBytes, '\n')); err != nil {
		fmt.Fprintf(stderr, "OUTPUT_ERROR: %v\n", err)
		return exitUsage
	}
	fmt.Fprintf(stdout, "%s: %s\n", report.Overall, filepath.Clean(*outputPath))
	switch report.Overall {
	case "PASS":
		return exitPass
	case "FAIL":
		return exitFail
	case "INCONCLUSIVE":
		return exitInconclusive
	default:
		return exitUsage
	}
}

func readBounded(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if st.Size() > maxInputBytes {
		return nil, fmt.Errorf("input exceeds %d-byte limit", maxInputBytes)
	}
	data, err := io.ReadAll(io.LimitReader(f, maxInputBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxInputBytes {
		return nil, fmt.Errorf("input exceeds %d-byte limit", maxInputBytes)
	}
	return data, nil
}

func writeNew(path string, data []byte) error {
	parent := filepath.Dir(filepath.Clean(path))
	if err := os.MkdirAll(parent, 0700); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(data); err != nil {
		return err
	}
	return f.Sync()
}
