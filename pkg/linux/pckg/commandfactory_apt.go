// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package pckg

type APTCommandFactory struct{}

func (c APTCommandFactory) NewCleanCacheCmd() (cmd, capabilities []string) {
	cmd = []string{"apt", "--quiet", "clean"}
	capabilities = []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_FOWNER",
	}
	return
}

func (c APTCommandFactory) NewInstallCmd(packages []string) (cmd, capabilities []string) {
	cmd = []string{"apt", "--quiet", "--yes", "install"}
	cmd = append(cmd, packages...)
	capabilities = []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_FOWNER",
		"CAP_SETGID",
		"CAP_SETUID",
	}
	return
}

func (c APTCommandFactory) NewListInstalledPackagesCmd() (
	cmd []string,
	capabilities []string,
	parse func([]string) ([]string, error),
) {
	cmd = []string{"apt-cache", "pkgnames"}
	capabilities = []string{}

	// expected line format: name
	parse = func(lines []string) ([]string, error) {
		return lines, nil
	}

	return
}

func (c APTCommandFactory) NewUpdateIndexCmd() (cmd, capabilities []string) {
	cmd = []string{"apt", "--quiet", "update"}
	capabilities = []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_FOWNER",
		"CAP_SETGID",
		"CAP_SETUID",
	}
	return
}

func (c APTCommandFactory) NewUpgradeCmd() (cmd, capabilities []string) {
	cmd = []string{"apt", "--quiet", "--yes", "upgrade"}
	capabilities = []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_FOWNER",
		"CAP_SETGID",
		"CAP_SETUID",
	}
	return
}

func (c APTCommandFactory) PackageManager() Manager {
	return APT
}
