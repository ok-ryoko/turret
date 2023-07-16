// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package packagemanager

type APKPackageManager struct{}

func (p APKPackageManager) NewCleanCacheCmd() (cmd, capabilities []string) {
	return []string{}, []string{}
}

func (p APKPackageManager) NewInstallCmd(packages []string) (cmd, capabilities []string) {
	cmd = []string{"apk", "--no-cache", "--no-progress", "add"}
	cmd = append(cmd, packages...)
	capabilities = []string{}
	return
}

func (p APKPackageManager) NewListInstalledPackagesCmd() (cmd, capabilities []string) {
	cmd = []string{"apk", "list", "--installed"}
	capabilities = []string{}
	return
}

func (p APKPackageManager) NewUpdateIndexCmd() (cmd, capabilities []string) {
	return []string{}, []string{}
}

func (p APKPackageManager) NewUpgradeCmd() (cmd, capabilities []string) {
	cmd = []string{"apk", "--no-cache", "--no-progress", "upgrade"}
	capabilities = []string{}
	return
}

func (p APKPackageManager) PackageManager() PackageManager {
	return APK
}
