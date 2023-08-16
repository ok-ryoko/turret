package pckg

import (
	"os"
	"strings"
	"testing"
)

func TestParseXBPSPackages(t *testing.T) {
	cf := XBPSCommandFactory{}
	_, _, parse := cf.NewListInstalledPackagesCmd()

	raw, err := os.ReadFile("testdata/xbps.txt")
	if err != nil {
		t.Fatalf("reading test data: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	actual, err := parse(lines)
	if err != nil {
		t.Fatalf("parsing packages from test data: %v", err)
	}

	expected := []string{
		"acl",
		"attr",
		"base-files",
		"base-minimal",
		"bzip2",
		"ca-certificates",
		"coreutils",
		"dash",
		"diffutils",
		"eudev-libudev",
		"findutils",
		"gawk",
		"glibc",
		"glibc-locales",
		"gmp",
		"grep",
		"gzip",
		"iana-etc",
		"libarchive",
		"libblkid",
		"libcap",
		"libcap-ng",
		"libcrypto1.1",
		"libdb",
		"libfdisk",
		"liblz4",
		"liblzma",
		"libmount",
		"libpcre2",
		"libreadline8",
		"libsmartcols",
		"libssl1.1",
		"libuuid",
		"libxbps",
		"libzstd",
		"ncurses-libs",
		"nvi",
		"openssl",
		"pam",
		"pam-base",
		"pam-libs",
		"procps-ng",
		"removed-packages",
		"run-parts",
		"runit",
		"runit-void",
		"sed",
		"shadow",
		"tar",
		"tzdata",
		"util-linux",
		"util-linux-common",
		"which",
		"xbps",
		"xbps-triggers",
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
