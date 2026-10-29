package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func dataset(provenance string, reference, observed []map[string]any) map[string]any {
	return map[string]any{"reference": reference, "observed": observed, "coverage": map[string]any{"ranges": []map[string]any{{"from": 100, "to": 101}}, "provenance": provenance, "completeness": "complete"}}
}
func event(tx string, order int, data int) map[string]any {
	return map[string]any{"ledgerSeq": 100, "txHash": tx, "txApplicationOrder": order, "eventIndex": 0, "topics": []any{"transfer", "A", "B"}, "data": data}
}
func writeInput(t *testing.T, dir string, value any) string {
	t.Helper()
	p := filepath.Join(dir, "input.json")
	b, _ := json.Marshal(value)
	if err := os.WriteFile(p, b, 0600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestRunPass(t *testing.T) {
	d := t.TempDir()
	in := writeInput(t, d, dataset("captured-final-ledger-export", []map[string]any{event("a", 0, 10)}, []map[string]any{event("a", 0, 10)}))
	out := filepath.Join(d, "report.json")
	if code := run([]string{"--input", in, "--output", out}, &strings.Builder{}, &strings.Builder{}); code != exitPass {
		t.Fatalf("code=%d", code)
	}
	if _, err := os.Stat(out); err != nil {
		t.Fatal(err)
	}
}
func TestRunFailAndSafeWrite(t *testing.T) {
	d := t.TempDir()
	in := writeInput(t, d, dataset("captured-final-ledger-export", []map[string]any{event("a", 0, 10), event("b", 1, 5)}, []map[string]any{event("a", 0, 10), event("a", 0, 10)}))
	out := filepath.Join(d, "report.json")
	if code := run([]string{"--input", in, "--output", out}, &strings.Builder{}, &strings.Builder{}); code != exitFail {
		t.Fatalf("code=%d", code)
	}
	if code := run([]string{"--input", in, "--output", out}, &strings.Builder{}, &strings.Builder{}); code != exitUsage {
		t.Fatalf("collision code=%d", code)
	}
}
func TestRunInconclusive(t *testing.T) {
	d := t.TempDir()
	in := writeInput(t, d, dataset("unknown", []map[string]any{event("a", 0, 10)}, []map[string]any{event("a", 0, 10)}))
	out := filepath.Join(d, "report.json")
	if code := run([]string{"--input", in, "--output", out}, &strings.Builder{}, &strings.Builder{}); code != exitInconclusive {
		t.Fatalf("code=%d", code)
	}
}
func TestRunRejectsMalformedJSON(t *testing.T) {
	d := t.TempDir()
	in := filepath.Join(d, "bad.json")
	if err := os.WriteFile(in, []byte("{"), 0600); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(d, "report.json")
	if code := run([]string{"--input", in, "--output", out}, &strings.Builder{}, &strings.Builder{}); code != exitUsage {
		t.Fatalf("code=%d", code)
	}
}

func TestRunRejectsTrailingJSON(t *testing.T) {
	d := t.TempDir()
	in := filepath.Join(d, "bad.json")
	valid := dataset("captured-final-ledger-export", []map[string]any{event("a", 0, 10)}, []map[string]any{event("a", 0, 10)})
	b, _ := json.Marshal(valid)
	if err := os.WriteFile(in, append(append(b, '\n'), []byte(`{"extra":true}`)...), 0600); err != nil {
		t.Fatal(err)
	}
	if code := run([]string{"--input", in, "--output", filepath.Join(d, "report.json")}, &strings.Builder{}, &strings.Builder{}); code != exitUsage {
		t.Fatalf("code=%d", code)
	}
}
