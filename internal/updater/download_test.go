package updater

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestSafeArchivePathRejectsTraversal(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"../escape", "/absolute", "a/../../escape"} {
		if _, err := safeArchivePath(dir, name); err == nil {
			t.Fatalf("expected unsafe path %q to fail", name)
		}
	}
}

func TestExtractZipAcceptsExpectedFiles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "release.zip")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	for name, content := range map[string]string{"cmdr.exe": "binary", "README.md": "readme", "LICENSE": "license"} {
		entry, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	binary, err := ExtractArchive(path, t.TempDir(), "release.zip", "windows")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(binary)
	if err != nil || string(data) != "binary" {
		t.Fatalf("binary = %q, %v", data, err)
	}
}

func TestExtractZipRejectsUnexpectedAndTraversal(t *testing.T) {
	for _, name := range []string{"../cmdr.exe", "evil.txt"} {
		path := filepath.Join(t.TempDir(), "release.zip")
		file, _ := os.Create(path)
		writer := zip.NewWriter(file)
		entry, _ := writer.Create(name)
		_, _ = entry.Write([]byte("data"))
		_ = writer.Close()
		_ = file.Close()
		if _, err := ExtractArchive(path, t.TempDir(), "release.zip", "windows"); err == nil {
			t.Fatalf("expected %q to fail", name)
		}
	}
}

func TestExtractTarGzAcceptsExpectedFiles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "release.tar.gz")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(file)
	tw := tar.NewWriter(gz)
	for _, name := range []string{"cmdr", "README.md", "LICENSE"} {
		content := "data"
		if name == "cmdr" {
			content = "binary"
		}
		header := &tar.Header{Name: name, Mode: 0755, Size: int64(len(content)), Typeflag: tar.TypeReg}
		if err := tw.WriteHeader(header); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	binary, err := ExtractArchive(path, t.TempDir(), "release.tar.gz", "linux")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(binary)
	if err != nil || string(data) != "binary" {
		t.Fatalf("binary = %q, %v", data, err)
	}
}

func TestDownloadToFileHonorsHTTPStatusAndSize(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()
	if err := DownloadToFile(server.URL, filepath.Join(t.TempDir(), "archive"), 1024); err == nil {
		t.Fatal("expected HTTP 403 error")
	}
}
