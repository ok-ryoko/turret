// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package packagemanager

type ZypperPackageManager struct{}

func (p ZypperPackageManager) NewCleanCacheCmd() (cmd, capabilities []string) {
	cmd = []string{"zypper", "--non-interactive", "clean", "--all"}
	capabilities = []string{}
	return
}

func (p ZypperPackageManager) NewInstallCmd(packages []string) (cmd, capabilities []string) {
	cmd = []string{"zypper", "--non-interactive", "install", "--no-recommends"}
	cmd = append(cmd, packages...)
	capabilities = []string{}
	return
}

func (p ZypperPackageManager) NewListInstalledPackagesCmd() (cmd, capabilities []string) {
	cmd = []string{"zypper", "search", "--installed-only"}
	capabilities = []string{}
	return
}

func (p ZypperPackageManager) NewUpdateIndexCmd() (cmd, capabilities []string) {
	return []string{}, []string{}
}

func (p ZypperPackageManager) NewUpgradeCmd() (cmd, capabilities []string) {
	cmd = []string{"zypper", "--non-interactive", "patch"}
	capabilities = []string{}
	return
}

func (p ZypperPackageManager) PackageManager() PackageManager {
	return Zypper
}
