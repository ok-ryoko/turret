// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package pckg

type PacmanCommandFactory struct{}

func (f PacmanCommandFactory) NewCleanCacheCmd() (cmd, capabilities []string) {
	cmd = []string{"pacman", "--sync", "--clean", "--clean", "--noconfirm", "--quiet"}
	return cmd, []string{}
}

func (f PacmanCommandFactory) NewInstallCmd(packages []string) (cmd, capabilities []string) {
	cmd = []string{"pacman", "--sync", "--noconfirm", "--noprogressbar", "--quiet"}
	cmd = append(cmd, packages...)
	capabilities = []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_FOWNER",
		"CAP_SYS_CHROOT",
	}
	return cmd, capabilities
}

func (f PacmanCommandFactory) NewListInstalledPackagesCmd() (
	cmd []string,
	capabilities []string,
	parse func([]string) ([]string, error),
) {
	cmd = []string{
		"pacman",
		"--color", "never",
		"--query",
		"--quiet",
	}

	// expected line format: name
	parse = func(lines []string) ([]string, error) {
		return lines, nil
	}

	return cmd, []string{}, parse
}

func (f PacmanCommandFactory) NewUpdateIndexCmd() (cmd, capabilities []string) {
	return []string{}, []string{}
}

func (f PacmanCommandFactory) NewUpgradeCmd() (cmd, capabilities []string) {
	cmd = []string{"pacman", "--sync", "--sysupgrade", "--refresh", "--noconfirm", "--noprogressbar", "--quiet"}
	capabilities = []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_FOWNER",
		"CAP_SYS_CHROOT",
	}
	return cmd, capabilities
}

func (f PacmanCommandFactory) PackageManager() Manager {
	return Pacman
}
