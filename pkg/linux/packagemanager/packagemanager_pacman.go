// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package packagemanager

type PacmanPackageManager struct{}

func (p PacmanPackageManager) NewCleanCacheCmd() (cmd, capabilities []string) {
	cmd = []string{"pacman", "--sync", "--clean", "--clean", "--noconfirm"}
	capabilities = []string{}
	return
}

func (p PacmanPackageManager) NewInstallCmd(packages []string) (cmd, capabilities []string) {
	cmd = []string{"pacman", "--sync", "--noconfirm", "--noprogressbar"}
	cmd = append(cmd, packages...)
	capabilities = []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_FOWNER",
		"CAP_SYS_CHROOT",
	}
	return
}

func (p PacmanPackageManager) NewListInstalledPackagesCmd() (cmd, capabilities []string) {
	cmd = []string{"pacman", "--query"}
	capabilities = []string{}
	return
}

func (p PacmanPackageManager) NewUpdateIndexCmd() (cmd, capabilities []string) {
	return []string{}, []string{}
}

func (p PacmanPackageManager) NewUpgradeCmd() (cmd, capabilities []string) {
	cmd = []string{"pacman", "--sync", "--sysupgrade", "--refresh", "--noconfirm", "--noprogressbar"}
	capabilities = []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_FOWNER",
		"CAP_SYS_CHROOT",
	}
	return
}

func (p PacmanPackageManager) PackageManager() PackageManager {
	return Pacman
}
