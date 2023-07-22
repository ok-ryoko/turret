// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package pckg

type ZypperCommandFactory struct{}

func (c ZypperCommandFactory) NewCleanCacheCmd() (cmd, capabilities []string) {
	cmd = []string{"zypper", "--non-interactive", "clean", "--all"}
	capabilities = []string{}
	return
}

func (c ZypperCommandFactory) NewInstallCmd(packages []string) (cmd, capabilities []string) {
	cmd = []string{"zypper", "--non-interactive", "install", "--no-recommends"}
	cmd = append(cmd, packages...)
	capabilities = []string{}
	return
}

func (c ZypperCommandFactory) NewListInstalledPackagesCmd() (cmd, capabilities []string) {
	cmd = []string{"zypper", "search", "--installed-only"}
	capabilities = []string{}
	return
}

func (c ZypperCommandFactory) NewUpdateIndexCmd() (cmd, capabilities []string) {
	return []string{}, []string{}
}

func (c ZypperCommandFactory) NewUpgradeCmd() (cmd, capabilities []string) {
	cmd = []string{"zypper", "--non-interactive", "patch"}
	capabilities = []string{}
	return
}

func (c ZypperCommandFactory) PackageManager() Manager {
	return Zypper
}
