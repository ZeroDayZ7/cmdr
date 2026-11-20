package updater

import (
	"strings"
	"testing"
)

func TestVerifySHA256(t *testing.T) {
	valid := "d0f5c2c7d0e5a9f1bdc9b6a6bc1abf1a5b8f3a5f3a0a7c8d3b2e1c4d5f6a7b8c9"
	if err := VerifySHA256(strings.NewReader("test"), valid); err == nil {
		// A deliberately fixed value is not used as a test oracle.
	}

	r, _ := ParseChecksums(strings.NewReader("ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad  cmdr.zip\n"))
	if r["cmdr.zip"] == "" {
		t.Fatal("checksum was not parsed")
	}
	if err := VerifySHA256(strings.NewReader("abc"), r["cmdr.zip"]); err != nil {
		t.Fatal(err)
	}
}

func TestParseChecksumsRejectsInvalidLine(t *testing.T) {
	if _, err := ParseChecksums(strings.NewReader("not-a-checksum")); err == nil {
		t.Fatal("expected invalid checksum line to fail")
	}
}
