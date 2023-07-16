// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package packagemanager

type APTPackageManager struct{}

func (p APTPackageManager) NewCleanCacheCmd() (cmd, capabilities []string) {
	cmd = []string{"apt", "--quiet", "clean"}
	capabilities = []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_FOWNER",
	}
	return
}

func (p APTPackageManager) NewInstallCmd(packages []string) (cmd, capabilities []string) {
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

func (p APTPackageManager) NewListInstalledPackagesCmd() (cmd, capabilities []string) {
	cmd = []string{"apt", "list"}
	capabilities = []string{}
	return
}

func (p APTPackageManager) NewUpdateIndexCmd() (cmd, capabilities []string) {
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

func (p APTPackageManager) NewUpgradeCmd() (cmd, capabilities []string) {
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

func (p APTPackageManager) PackageManager() PackageManager {
	return APT
}
