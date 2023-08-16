package pckg

import (
	"os"
	"strings"
	"testing"
)

func TestParseAPTPackages(t *testing.T) {
	cf := APTCommandFactory{}
	_, _, parse := cf.NewListInstalledPackagesCmd()

	raw, err := os.ReadFile("testdata/apt.txt")
	if err != nil {
		t.Fatalf("reading test data: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	actual, err := parse(lines)
	if err != nil {
		t.Fatalf("parsing packages from test data: %v", err)
	}

	expected := []string{
		"dpkg",
		"libsmartcols1",
		"libpam-runtime",
		"coreutils",
		"libext2fs2",
		"libffi8",
		"libaudit-common",
		"apt",
		"libnettle8",
		"libacl1",
		"libtinfo6",
		"tzdata",
		"libunistring2",
		"libidn2-0",
		"libtasn1-6",
		"libselinux1",
		"liblzma5",
		"sed",
		"tar",
		"base-passwd",
		"libdb5.3",
		"libp11-kit0",
		"libapt-pkg6.0",
		"debconf",
		"libsystemd0",
		"libmount1",
		"debianutils",
		"libsepol2",
		"libzstd1",
		"util-linux",
		"util-linux-extra",
		"libudev1",
		"libgpg-error0",
		"usr-is-merged",
		"debian-archive-keyring",
		"libcom-err2",
		"diffutils",
		"libcap2",
		"libc6",
		"libpam-modules",
		"libuuid1",
		"gpgv",
		"bash",
		"libgcrypt20",
		"grep",
		"mawk",
		"libcap-ng0",
		"libsemanage2",
		"base-files",
		"ncurses-base",
		"gzip",
		"login",
		"hostname",
		"adduser",
		"findutils",
		"libgmp10",
		"libxxhash0",
		"libpcre2-8-0",
		"libcrypt1",
		"libpam-modules-bin",
		"init-system-helpers",
		"libsemanage-common",
		"libseccomp2",
		"mount",
		"perl-base",
		"libpam0g",
		"libc-bin",
		"libattr1",
		"libaudit1",
		"libhogweed6",
		"libblkid1",
		"libgnutls30",
		"sysvinit-utils",
		"logsave",
		"liblz4-1",
		"ncurses-bin",
		"libbz2-1.0",
		"zlib1g",
		"gcc-12-base",
		"libmd0",
		"dash",
		"bsdutils",
		"libss2",
		"libstdc++6",
		"libdebconfclient0",
		"e2fsprogs",
		"passwd",
		"libgcc-s1",
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
