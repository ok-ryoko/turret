// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package pckg

type APTCommandFactory struct{}

func (f APTCommandFactory) NewCleanCacheCmd() (cmd, capabilities []string) {
	cmd = []string{"apt", "--quiet", "clean"}
	capabilities = []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_FOWNER",
	}
	return cmd, capabilities
}

func (f APTCommandFactory) NewInstallCmd(packages []string) (cmd, capabilities []string) {
	cmd = []string{"apt", "--quiet", "--yes", "install"}
	cmd = append(cmd, packages...)
	capabilities = []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_FOWNER",
		"CAP_SETGID",
		"CAP_SETUID",
	}
	return cmd, capabilities
}

func (f APTCommandFactory) NewListInstalledPackagesCmd() (
	cmd []string,
	capabilities []string,
	parse func([]string) ([]string, error),
) {
	cmd = []string{"apt-cache", "pkgnames"}

	// expected line format: name
	parse = func(lines []string) ([]string, error) {
		return lines, nil
	}

	return cmd, []string{}, parse
}

func (f APTCommandFactory) NewUpdateIndexCmd() (cmd, capabilities []string) {
	cmd = []string{"apt", "--quiet", "update"}
	capabilities = []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_FOWNER",
		"CAP_SETGID",
		"CAP_SETUID",
	}
	return cmd, capabilities
}

func (f APTCommandFactory) NewUpgradeCmd() (cmd, capabilities []string) {
	cmd = []string{"apt", "--quiet", "--yes", "upgrade"}
	capabilities = []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_FOWNER",
		"CAP_SETGID",
		"CAP_SETUID",
	}
	return cmd, capabilities
}

func (f APTCommandFactory) Backend() Backend {
	return APT
}
