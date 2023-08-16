// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package pckg

type PacmanCommandFactory struct{}

func (c PacmanCommandFactory) NewCleanCacheCmd() (cmd, capabilities []string) {
	cmd = []string{"pacman", "--sync", "--clean", "--clean", "--noconfirm", "--quiet"}
	capabilities = []string{}
	return
}

func (c PacmanCommandFactory) NewInstallCmd(packages []string) (cmd, capabilities []string) {
	cmd = []string{"pacman", "--sync", "--noconfirm", "--noprogressbar", "--quiet"}
	cmd = append(cmd, packages...)
	capabilities = []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_FOWNER",
		"CAP_SYS_CHROOT",
	}
	return
}

func (c PacmanCommandFactory) NewListInstalledPackagesCmd() (
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

	capabilities = []string{}

	// expected line format: name
	parse = func(lines []string) ([]string, error) {
		return lines, nil
	}

	return
}

func (c PacmanCommandFactory) NewUpdateIndexCmd() (cmd, capabilities []string) {
	return []string{}, []string{}
}

func (c PacmanCommandFactory) NewUpgradeCmd() (cmd, capabilities []string) {
	cmd = []string{"pacman", "--sync", "--sysupgrade", "--refresh", "--noconfirm", "--noprogressbar", "--quiet"}
	capabilities = []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_FOWNER",
		"CAP_SYS_CHROOT",
	}
	return
}

func (c PacmanCommandFactory) PackageManager() Manager {
	return Pacman
}
