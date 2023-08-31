// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package pckg

import (
	"fmt"
	"strings"
)

type DNFCommandFactory struct{}

func (c DNFCommandFactory) NewCleanCacheCmd() (cmd, capabilities []string) {
	cmd = []string{"dnf", "--quiet", "clean", "all"}
	return cmd, []string{}
}

func (c DNFCommandFactory) NewInstallCmd(packages []string) (cmd, capabilities []string) {
	cmd = []string{"dnf", "--assumeyes", "--quiet", "--setopt=install_weak_deps=False", "install"}
	cmd = append(cmd, packages...)
	capabilities = []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_SETFCAP",
	}
	return cmd, capabilities
}

func (c DNFCommandFactory) NewListInstalledPackagesCmd() (
	cmd []string,
	capabilities []string,
	parse func([]string) ([]string, error),
) {
	cmd = []string{
		"dnf",
		"--color=never",
		"--quiet",
		"list",
		"--installed",
	}

	// expected line format: name.arch version repo
	parse = func(lines []string) ([]string, error) {
		if len(lines) < 2 {
			return []string{}, nil
		}
		result := make([]string, 0, len(lines)-1)
		for _, l := range lines[1:] {
			pkg, _, ok := strings.Cut(l, " ")
			if !ok {
				return nil, fmt.Errorf("expected space delimiter in line %q", l)
			}
			i := strings.LastIndex(pkg, ".")
			if i == -1 {
				return nil, fmt.Errorf("expected format 'name.arch' for field %q", pkg)
			}
			name := pkg[:i]
			result = append(result, name)
		}
		return result, nil
	}

	return cmd, []string{}, parse
}

func (c DNFCommandFactory) NewUpdateIndexCmd() (cmd, capabilities []string) {
	return []string{}, []string{}
}

func (c DNFCommandFactory) NewUpgradeCmd() (cmd, capabilities []string) {
	cmd = []string{"dnf", "--assumeyes", "--quiet", "--refresh", "upgrade"}
	capabilities = []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_SETFCAP",
	}
	return cmd, capabilities
}

func (c DNFCommandFactory) PackageManager() Manager {
	return DNF
}
