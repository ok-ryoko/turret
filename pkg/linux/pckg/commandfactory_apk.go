// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package pckg

type APKCommandFactory struct{}

func (c APKCommandFactory) NewCleanCacheCmd() (cmd, capabilities []string) {
	return []string{}, []string{}
}

func (c APKCommandFactory) NewInstallCmd(packages []string) (cmd, capabilities []string) {
	cmd = []string{"apk", "--no-cache", "--no-progress", "--quiet", "add"}
	cmd = append(cmd, packages...)
	capabilities = []string{}
	return
}

func (c APKCommandFactory) NewListInstalledPackagesCmd() (cmd, capabilities []string) {
	cmd = []string{"apk", "list", "--installed"}
	capabilities = []string{}
	return
}

func (c APKCommandFactory) NewUpdateIndexCmd() (cmd, capabilities []string) {
	return []string{}, []string{}
}

func (c APKCommandFactory) NewUpgradeCmd() (cmd, capabilities []string) {
	cmd = []string{"apk", "--no-cache", "--no-progress", "--quiet", "upgrade"}
	capabilities = []string{}
	return
}

func (c APKCommandFactory) PackageManager() Manager {
	return APK
}
