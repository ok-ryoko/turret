// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package pckg

type DNFCommandFactory struct{}

func (c DNFCommandFactory) NewCleanCacheCmd() (cmd, capabilities []string) {
	cmd = []string{"dnf", "clean", "all"}
	capabilities = []string{}
	return
}

func (c DNFCommandFactory) NewInstallCmd(packages []string) (cmd, capabilities []string) {
	cmd = []string{"dnf", "--assumeyes", "--setopt=install_weak_deps=False", "install"}
	cmd = append(cmd, packages...)
	capabilities = []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_SETFCAP",
	}
	return
}

func (c DNFCommandFactory) NewListInstalledPackagesCmd() (cmd, capabilities []string) {
	cmd = []string{"dnf", "list", "--installed"}
	capabilities = []string{}
	return
}

func (c DNFCommandFactory) NewUpdateIndexCmd() (cmd, capabilities []string) {
	return []string{}, []string{}
}

func (c DNFCommandFactory) NewUpgradeCmd() (cmd, capabilities []string) {
	cmd = []string{"dnf", "--assumeyes", "--refresh", "upgrade"}
	capabilities = []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_SETFCAP",
	}
	return
}

func (c DNFCommandFactory) PackageManager() Manager {
	return DNF
}
