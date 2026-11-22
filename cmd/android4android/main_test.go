package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"android4android/internal/preflight"
)

func TestDoctorFailureStillPrintsJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	check := func() preflight.Report {
		return preflight.Report{OS: "linux", Arch: "amd64", Checks: []preflight.Result{
			{ID: "kvm", Status: preflight.Fail, Detail: "unavailable"},
		}}
	}
	if code := run([]string{"doctor", "--json"}, &stdout, &stderr, check); code != 2 {
		t.Fatalf("code = %d, want 2", code)
	}
	var got preflight.Report
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil || len(got.Checks) != 1 || got.Checks[0].ID != "kvm" {
		t.Fatalf("got %s (%v)", stdout.String(), err)
	}
}

func TestUsageDoesNotRunDiagnostic(t *testing.T) {
	for _, tc := range []struct {
		args []string
		code int
	}{
		{[]string{"--help"}, 0},
		{[]string{"launch"}, 64},
		{[]string{"doctor", "invalid"}, 64},
	} {
		var stdout, stderr bytes.Buffer
		code := run(tc.args, &stdout, &stderr, func() preflight.Report {
			t.Fatal("diagnostic ran")
			return preflight.Report{}
		})
		if code != tc.code || !strings.Contains(stdout.String()+stderr.String(), "Usage:") {
			t.Fatalf("args %q: code %d, output %q", tc.args, code, stdout.String()+stderr.String())
		}
	}
}
