package pckg

import (
	"os"
	"strings"
	"testing"
)

func TestParseAPKPackages(t *testing.T) {
	cf := APKCommandFactory{}
	_, _, parse := cf.NewListInstalledPackagesCmd()

	raw, err := os.ReadFile("testdata/apk.txt")
	if err != nil {
		t.Fatalf("reading test data: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	actual, err := parse(lines)
	if err != nil {
		t.Fatalf("parsing packages from test data: %v", err)
	}

	expected := []string{
		"alpine-baselayout",
		"alpine-baselayout-data",
		"alpine-keys",
		"apk-tools",
		"busybox",
		"busybox-binsh",
		"ca-certificates-bundle",
		"libc-utils",
		"libcrypto3",
		"libssl3",
		"musl",
		"musl-utils",
		"scanelf",
		"ssl_client",
		"zlib",
	}

	if len(actual) != len(expected) {
		t.Fatalf("expected %d packages, found %d", len(expected), len(actual))
	}

	for i := range expected {
		if actual[i] != expected[i] {
			t.Errorf("expected package %s at position %d, found %s", expected[i], i, actual[i])
		}
	}
}
