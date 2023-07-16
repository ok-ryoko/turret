// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package packagemanager

type DNFPackageManager struct{}

func (p DNFPackageManager) NewCleanCacheCmd() (cmd, capabilities []string) {
	cmd = []string{"dnf", "clean", "all"}
	capabilities = []string{}
	return
}

func (p DNFPackageManager) NewInstallCmd(packages []string) (cmd, capabilities []string) {
	cmd = []string{"dnf", "--assumeyes", "--setopt=install_weak_deps=False", "install"}
	cmd = append(cmd, packages...)
	capabilities = []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_SETFCAP",
	}
	return
}

func (p DNFPackageManager) NewListInstalledPackagesCmd() (cmd, capabilities []string) {
	cmd = []string{"dnf", "list", "--installed"}
	capabilities = []string{}
	return
}

func (p DNFPackageManager) NewUpdateIndexCmd() (cmd, capabilities []string) {
	return []string{}, []string{}
}

func (p DNFPackageManager) NewUpgradeCmd() (cmd, capabilities []string) {
	cmd = []string{"dnf", "--assumeyes", "--refresh", "upgrade"}
	capabilities = []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_SETFCAP",
	}
	return
}

func (p DNFPackageManager) PackageManager() PackageManager {
	return DNF
}
