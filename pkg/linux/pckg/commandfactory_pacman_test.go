package pckg

import (
	"os"
	"strings"
	"testing"
)

func TestParsePacmanPackages(t *testing.T) {
	cf := PacmanCommandFactory{}
	_, _, parse := cf.NewListInstalledPackagesCmd()

	raw, err := os.ReadFile("testdata/pacman.txt")
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
		"archlinux-keyring",
		"argon2",
		"attr",
		"audit",
		"base",
		"bash",
		"brotli",
		"bzip2",
		"ca-certificates",
		"ca-certificates-mozilla",
		"ca-certificates-utils",
		"coreutils",
		"cryptsetup",
		"curl",
		"dbus",
		"device-mapper",
		"e2fsprogs",
		"expat",
		"file",
		"filesystem",
		"findutils",
		"gawk",
		"gcc-libs",
		"gdbm",
		"gettext",
		"glib2",
		"glibc",
		"gmp",
		"gnupg",
		"gnutls",
		"gpgme",
		"grep",
		"gzip",
		"hwdata",
		"iana-etc",
		"icu",
		"iproute2",
		"iptables",
		"iputils",
		"json-c",
		"kbd",
		"keyutils",
		"kmod",
		"krb5",
		"less",
		"libarchive",
		"libassuan",
		"libbpf",
		"libcap",
		"libcap-ng",
		"libelf",
		"libevent",
		"libffi",
		"libgcrypt",
		"libgpg-error",
		"libidn2",
		"libksba",
		"libldap",
		"libmnl",
		"libnetfilter_conntrack",
		"libnfnetlink",
		"libnftnl",
		"libnghttp2",
		"libnl",
		"libp11-kit",
		"libpcap",
		"libpsl",
		"libsasl",
		"libseccomp",
		"libsecret",
		"libssh2",
		"libsysprof-capture",
		"libtasn1",
		"libtirpc",
		"libunistring",
		"libutempter",
		"libverto",
		"libxcrypt",
		"libxml2",
		"licenses",
		"linux-api-headers",
		"lz4",
		"mpfr",
		"ncurses",
		"nettle",
		"npth",
		"openssl",
		"p11-kit",
		"pacman",
		"pacman-mirrorlist",
		"pam",
		"pambase",
		"pciutils",
		"pcre2",
		"pinentry",
		"popt",
		"procps-ng",
		"psmisc",
		"readline",
		"sed",
		"shadow",
		"sqlite",
		"systemd",
		"systemd-libs",
		"systemd-sysvcompat",
		"tar",
		"tpm2-tss",
		"tzdata",
		"util-linux",
		"util-linux-libs",
		"xz",
		"zlib",
		"zstd",
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
