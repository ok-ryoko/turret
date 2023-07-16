// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package packagemanager

type XBPSPackageManager struct{}

func (p XBPSPackageManager) NewCleanCacheCmd() (cmd, capabilities []string) {
	cmd = []string{"xbps-remove", "--clean-cache", "--yes"}
	capabilities = []string{}
	return
}

func (p XBPSPackageManager) NewInstallCmd(packages []string) (cmd, capabilities []string) {
	cmd = []string{"xbps-install", "--yes"}
	cmd = append(cmd, packages...)
	capabilities = []string{"CAP_DAC_OVERRIDE"}
	return
}

func (p XBPSPackageManager) NewListInstalledPackagesCmd() (cmd, capabilities []string) {
	cmd = []string{"xbps-query", "--list-pkgs"}
	capabilities = []string{}
	return
}

func (p XBPSPackageManager) NewUpdateIndexCmd() (cmd, capabilities []string) {
	return []string{}, []string{}
}

func (p XBPSPackageManager) NewUpgradeCmd() (cmd, capabilities []string) {
	cmd = []string{"xbps-install", "--sync", "--update", "--yes"}
	capabilities = []string{"CAP_DAC_OVERRIDE"}
	return
}

func (p XBPSPackageManager) PackageManager() PackageManager {
	return XBPS
}
