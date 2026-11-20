package doctor

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestEnvironmentChecks(t *testing.T) {
	results := environmentChecks()
	if len(results) != 2 {
		t.Fatalf("environment checks = %d", len(results))
	}
	for _, result := range results {
		if result.Status != OK {
			t.Fatalf("environment result %s = %s", result.Name, result.Status)
		}
	}
}

func TestExecutableChecks(t *testing.T) {
	results := executableChecks()
	if len(results) < 2 {
		t.Fatalf("unexpected executable checks: %#v", results)
	}
	if results[0].Status != OK {
		t.Fatalf("executable path status = %s", results[0].Status)
	}
}

func TestConfigChecks(t *testing.T) {
	results := configChecks()
	if len(results) == 0 {
		t.Fatal("expected configuration checks")
	}
	for _, result := range results {
		if result.Status != OK && result.Status != WARN && result.Status != FAIL {
			t.Fatalf("invalid status: %s", result.Status)
		}
	}
}

func TestNetworkCheck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	results := networkCheck(Options{HTTPClient: server.Client(), APIURL: server.URL})
	if results[0].Status != OK {
		t.Fatalf("network status = %s", results[0].Status)
	}
}

func TestNetworkCheckHTTPErrorAndTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	results := networkCheck(Options{HTTPClient: server.Client(), APIURL: server.URL})
	if results[0].Status != WARN {
		t.Fatalf("network 403 status = %s", results[0].Status)
	}

	client := &http.Client{Timeout: 20 * time.Millisecond}
	results = networkCheck(Options{HTTPClient: client, APIURL: "http://10.255.255.1"})
	if results[0].Status != WARN {
		t.Fatalf("timeout status = %s", results[0].Status)
	}
}

func TestWriteTextAndJSONSummary(t *testing.T) {
	report := Report{
		Results: []Result{
			{Section: "Environment", Name: "OS", Status: OK, Message: "windows/amd64"},
			{Section: "Network", Name: "GitHub", Status: WARN, Message: "offline"},
		},
		Summary: Summary{Checks: 2, Passed: 1, Warnings: 1, Failed: 0},
	}
	var text bytes.Buffer
	WriteText(&text, report, false)
	if !strings.Contains(text.String(), "Warnings: 1") {
		t.Fatalf("text summary missing: %s", text.String())
	}

	var out bytes.Buffer
	if err := WriteJSON(&out, report); err != nil {
		t.Fatal(err)
	}
	var decoded Report
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("JSON is invalid: %v", err)
	}
	if decoded.Summary.Checks != 2 {
		t.Fatalf("JSON summary = %#v", decoded.Summary)
	}
}

func TestDirectoryWritableTemporaryProbe(t *testing.T) {
	dir := t.TempDir()
	if !directoryWritable(dir) {
		t.Fatal("expected temporary directory to be writable")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".cmdr-doctor-") {
			t.Fatalf("temporary probe was not removed: %s", entry.Name())
		}
	}
}
