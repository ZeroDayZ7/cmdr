package updater

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseVersion(t *testing.T) {
	valid := []string{"2.1.4", "v2.1.5", "2.1.5-beta.1"}
	for _, raw := range valid {
		if _, _, err := ParseVersion(raw); err != nil {
			t.Fatalf("ParseVersion(%q): %v", raw, err)
		}
	}
	invalid := []string{"", "v.1.0.1", "1.2", "01.2.3", "1.x.3"}
	for _, raw := range invalid {
		if _, _, err := ParseVersion(raw); err == nil {
			t.Fatalf("expected ParseVersion(%q) to fail", raw)
		}
	}
}

func TestCompareVersions(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		want int
	}{
		{"2.1.5", "2.1.4", 1},
		{"2.1.4", "2.1.5", -1},
		{"2.1.5", "2.1.5", 0},
		{"2.1.5", "2.1.5-beta.1", 1},
		{"2.1.6-rc.2", "2.1.6-rc.10", -1},
		{"2.1.6", "2.1.6-rc.2", 1},
		{"2.1.6-beta.1", "2.1.6-alpha.9", 1},
	} {
		got, err := CompareVersions(tc.a, tc.b)
		if err != nil {
			t.Fatal(err)
		}
		if got != tc.want {
			t.Fatalf("CompareVersions(%q,%q)=%d want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestStableVersionTag(t *testing.T) {
	for _, raw := range []string{"v.1.0.1", "v1.0", "v1.0.0-beta.1"} {
		if _, err := StableVersionTag(raw); err == nil {
			t.Fatalf("expected %q to fail", raw)
		}
	}
	version, err := StableVersionTag("v2.1.5")
	if err != nil || version != "2.1.5" {
		t.Fatalf("unexpected stable tag result: %q, %v", version, err)
	}
}

func TestFetchLatestRelease(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"tag_name":"v2.1.5","draft":false,"prerelease":false,"assets":[{"name":"cmdr_2.1.5_linux_amd64.tar.gz","browser_download_url":"https://example.test/cmdr.tar.gz","size":100}]}`))
	}))
	defer server.Close()

	release, err := FetchLatestRelease(server.URL, false)
	if err != nil {
		t.Fatal(err)
	}
	if release.TagName != "v2.1.5" {
		t.Fatalf("tag = %q", release.TagName)
	}
}

func TestFetchLatestReleaseRejectsDraftPrereleaseAndHTTPError(t *testing.T) {
	for _, body := range []string{
		`{"tag_name":"v2.1.5","draft":true,"prerelease":false}`,
		`{"tag_name":"v2.1.5","draft":false,"prerelease":true}`,
	} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(body))
		}))
		release, err := FetchLatestRelease(server.URL, false)
		server.Close()
		if err == nil || release != nil {
			t.Fatalf("expected draft/prerelease rejection")
		}
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()
	if _, err := FetchLatestRelease(server.URL, false); err == nil {
		t.Fatal("expected HTTP 403 error")
	}
}

func TestPlatformAssetName(t *testing.T) {
	if name, err := PlatformAssetName("windows", "amd64", "2.1.5"); err != nil || name != "cmdr_2.1.5_windows_amd64.zip" {
		t.Fatalf("windows name: %q, %v", name, err)
	}
	if name, err := PlatformAssetName("linux", "amd64", "2.1.5"); err != nil || name != "cmdr_2.1.5_linux_amd64.tar.gz" {
		t.Fatalf("linux name: %q, %v", name, err)
	}
	if _, err := PlatformAssetName("darwin", "amd64", "2.1.5"); err == nil {
		t.Fatal("expected unsupported platform")
	}
}
