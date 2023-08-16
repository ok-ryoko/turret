package pckg

import (
	"os"
	"strings"
	"testing"
)

func TestParseZypperPackages(t *testing.T) {
	cf := ZypperCommandFactory{}
	_, _, parse := cf.NewListInstalledPackagesCmd()

	raw, err := os.ReadFile("testdata/zypper.txt")
	if err != nil {
		t.Fatalf("reading test data: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	actual, err := parse(lines)
	if err != nil {
		t.Fatalf("parsing packages from test data: %v", err)
	}

	expected := []string{
		"aaa_base",
		"bash",
		"bash-sh",
		"boost-license1_82_0",
		"ca-certificates",
		"ca-certificates-mozilla",
		"chkstat",
		"coreutils",
		"cracklib-dict-small",
		"crypto-policies",
		"curl",
		"file-magic",
		"filesystem",
		"fillup",
		"findutils",
		"glibc",
		"glibc-locale-base",
		"gpg2",
		"grep",
		"gzip",
		"krb5",
		"libabsl2301_0_0",
		"libacl1",
		"libassuan0",
		"libattr1",
		"libaudit1",
		"libaugeas0",
		"libblkid1",
		"libboost_thread1_82_0",
		"libbrotlicommon1",
		"libbrotlidec1",
		"libbz2-1",
		"libcap-ng0",
		"libcap2",
		"libcom_err2",
		"libcrypt1",
		"libcurl4",
		"libeconf0",
		"libfa1",
		"libfdisk1",
		"libffi8",
		"libgcc_s1",
		"libgcrypt20",
		"libglib-2_0-0",
		"libgmp10",
		"libgpg-error0",
		"libgpgme11",
		"libidn2-0",
		"libkeyutils1",
		"libksba8",
		"libldap2",
		"liblua5_4-5",
		"liblz4-1",
		"liblzma5",
		"libmagic1",
		"libmount1",
		"libncurses6",
		"libnghttp2-14",
		"libnpth0",
		"libnss_usrfiles2",
		"libopenssl3",
		"libp11-kit0",
		"libpcre2-8-0",
		"libpopt0",
		"libprocps8",
		"libprotobuf-lite23_4_0",
		"libproxy1",
		"libpsl5",
		"libreadline8",
		"libsasl2-3",
		"libselinux1",
		"libsemanage-conf",
		"libsemanage2",
		"libsepol2",
		"libsigc-2_0-0",
		"libsmartcols1",
		"libsolv-tools",
		"libsqlite3-0",
		"libssh-config",
		"libssh4",
		"libstdc++6",
		"libsubid4",
		"libsystemd0",
		"libtasn1-6",
		"libudev1",
		"libunistring5",
		"libusb-1_0-0",
		"libutempter0",
		"libuuid1",
		"libverto1",
		"libxml2-2",
		"libyaml-cpp0_7",
		"libz1",
		"libzck1",
		"libzstd1",
		"libzypp",
		"login_defs",
		"lsb-release",
		"ncurses-utils",
		"netcfg",
		"openssl",
		"openssl-3",
		"openSUSE-build-key",
		"openSUSE-release",
		"openSUSE-release-appliance-docker",
		"p11-kit",
		"p11-kit-tools",
		"pam",
		"patterns-base-fips",
		"permissions",
		"permissions-config",
		"pinentry",
		"procps",
		"rpm",
		"rpm-config-SUSE",
		"sed",
		"shadow",
		"system-group-hardware",
		"system-user-root",
		"sysuser-shadow",
		"tar",
		"terminfo-base",
		"timezone",
		"util-linux",
		"xz",
		"zypper",
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
