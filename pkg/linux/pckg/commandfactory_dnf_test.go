package pckg

import (
	"os"
	"strings"
	"testing"
)

func TestParseDNFPackages(t *testing.T) {
	cf := DNFCommandFactory{}
	_, _, parse := cf.NewListInstalledPackagesCmd()

	raw, err := os.ReadFile("testdata/dnf.txt")
	if err != nil {
		t.Fatalf("reading test data: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	actual, err := parse(lines)
	if err != nil {
		t.Fatalf("parsing packages from test data: %v", err)
	}

	expected := []string{
		"alternatives",
		"audit-libs",
		"authselect",
		"authselect-libs",
		"basesystem",
		"bash",
		"bzip2-libs",
		"ca-certificates",
		"coreutils",
		"coreutils-common",
		"cracklib",
		"crypto-policies",
		"curl",
		"cyrus-sasl-lib",
		"dnf",
		"dnf-data",
		"elfutils-default-yama-scope",
		"elfutils-libelf",
		"elfutils-libs",
		"expat",
		"fedora-gpg-keys",
		"fedora-release-common",
		"fedora-release-container",
		"fedora-release-identity-container",
		"fedora-repos",
		"file-libs",
		"filesystem",
		"findutils",
		"gawk",
		"gdbm-libs",
		"glib2",
		"glibc",
		"glibc-common",
		"glibc-minimal-langpack",
		"gmp",
		"gnupg2",
		"gnutls",
		"gpgme",
		"grep",
		"gzip",
		"ima-evm-utils",
		"json-c",
		"keyutils-libs",
		"krb5-libs",
		"libacl",
		"libarchive",
		"libassuan",
		"libattr",
		"libb2",
		"libblkid",
		"libbrotli",
		"libcap",
		"libcap-ng",
		"libcom_err",
		"libcomps",
		"libcurl",
		"libdb",
		"libdnf",
		"libeconf",
		"libevent",
		"libffi",
		"libfsverity",
		"libgcc",
		"libgcrypt",
		"libgomp",
		"libgpg-error",
		"libidn2",
		"libksba",
		"libmodulemd",
		"libmount",
		"libnghttp2",
		"libnsl2",
		"libpsl",
		"libpwquality",
		"librepo",
		"libreport-filesystem",
		"libselinux",
		"libsemanage",
		"libsepol",
		"libsigsegv",
		"libsmartcols",
		"libsolv",
		"libssh",
		"libssh-config",
		"libstdc++",
		"libtasn1",
		"libtirpc",
		"libunistring",
		"libuuid",
		"libverto",
		"libxcrypt",
		"libxml2",
		"libyaml",
		"libzstd",
		"lua-libs",
		"lz4-libs",
		"mpdecimal",
		"mpfr",
		"ncurses-base",
		"ncurses-libs",
		"nettle",
		"npth",
		"openldap",
		"openssl-libs",
		"p11-kit",
		"p11-kit-trust",
		"pam",
		"pam-libs",
		"pcre2",
		"pcre2-syntax",
		"popt",
		"publicsuffix-list-dafsa",
		"python-pip-wheel",
		"python3",
		"python3-dnf",
		"python3-gpg",
		"python3-hawkey",
		"python3-libcomps",
		"python3-libdnf",
		"python3-libs",
		"python3-rpm",
		"readline",
		"rootfiles",
		"rpm",
		"rpm-build-libs",
		"rpm-libs",
		"rpm-sequoia",
		"rpm-sign-libs",
		"sed",
		"setup",
		"shadow-utils",
		"sqlite-libs",
		"sudo",
		"systemd-libs",
		"tar",
		"tpm2-tss",
		"tzdata",
		"util-linux-core",
		"vim-data",
		"vim-minimal",
		"xz-libs",
		"yum",
		"zchunk-libs",
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
