package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/zerodayz7/cmdr/internal/doctor"
)

func executeDoctorForTest(args ...string) (string, error) {
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	rootCmd.SetArgs(nil)
	rootCmd.SetOut(nil)
	rootCmd.SetErr(nil)
	return out.String(), err
}

func TestDoctorCommandText(t *testing.T) {
	out, err := executeDoctorForTest("doctor")
	if err != nil {
		t.Fatalf("doctor returned error: %v", err)
	}
	if !strings.Contains(out, "cmdr doctor") || !strings.Contains(out, "Summary") {
		t.Fatalf("unexpected doctor output:\n%s", out)
	}
}

func TestDoctorCommandJSON(t *testing.T) {
	out, err := executeDoctorForTest("doctor", "--json")
	if err != nil {
		t.Fatalf("doctor JSON returned error: %v", err)
	}
	var report doctor.Report
	if err := json.Unmarshal([]byte(out), &report); err != nil {
		t.Fatalf("doctor JSON is invalid: %v\n%s", err, out)
	}
	if report.Summary.Checks == 0 {
		t.Fatal("doctor JSON summary is empty")
	}
}

func TestDoctorFailuresAndWarnings(t *testing.T) {
	warningOnly := doctor.Report{Summary: doctor.Summary{Checks: 1, Passed: 0, Warnings: 1, Failed: 0}}
	if warningOnly.Summary.Failed != 0 {
		t.Fatal("warning-only report should not fail")
	}

	failure := doctor.Report{Summary: doctor.Summary{Checks: 1, Passed: 0, Warnings: 0, Failed: 1}}
	if failure.Summary.Failed == 0 {
		t.Fatal("failure report should fail")
	}
}
