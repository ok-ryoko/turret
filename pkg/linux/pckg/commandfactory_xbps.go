// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package pckg

type XBPSCommandFactory struct{}

func (c XBPSCommandFactory) NewCleanCacheCmd() (cmd, capabilities []string) {
	cmd = []string{"xbps-remove", "--clean-cache", "--yes"}
	capabilities = []string{}
	return
}

func (c XBPSCommandFactory) NewInstallCmd(packages []string) (cmd, capabilities []string) {
	cmd = []string{"xbps-install", "--yes"}
	cmd = append(cmd, packages...)
	capabilities = []string{"CAP_DAC_OVERRIDE"}
	return
}

func (c XBPSCommandFactory) NewListInstalledPackagesCmd() (cmd, capabilities []string) {
	cmd = []string{"xbps-query", "--list-pkgs"}
	capabilities = []string{}
	return
}

func (c XBPSCommandFactory) NewUpdateIndexCmd() (cmd, capabilities []string) {
	return []string{}, []string{}
}

func (c XBPSCommandFactory) NewUpgradeCmd() (cmd, capabilities []string) {
	cmd = []string{"xbps-install", "--sync", "--update", "--yes"}
	capabilities = []string{"CAP_DAC_OVERRIDE"}
	return
}

func (c XBPSCommandFactory) PackageManager() Manager {
	return XBPS
}
