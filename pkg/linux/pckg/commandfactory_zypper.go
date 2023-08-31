// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package pckg

import (
	"fmt"
	"strings"
)

type ZypperCommandFactory struct{}

func (c ZypperCommandFactory) NewCleanCacheCmd() (cmd, capabilities []string) {
	cmd = []string{"zypper", "--non-interactive", "--quiet", "clean", "--all"}
	return cmd, []string{}
}

func (c ZypperCommandFactory) NewInstallCmd(packages []string) (cmd, capabilities []string) {
	cmd = []string{"zypper", "--non-interactive", "--quiet", "install", "--no-recommends"}
	cmd = append(cmd, packages...)
	return cmd, []string{}
}

func (c ZypperCommandFactory) NewListInstalledPackagesCmd() (
	cmd []string,
	capabilities []string,
	parse func([]string) ([]string, error),
) {
	cmd = []string{
		"zypper",
		"--non-interactive",
		"--quiet",
		"--terse",
		"--no-remote",
		"packages",
		"--installed-only",
	}

	// expected line format: status | repo | name | version | arch
	parse = func(lines []string) ([]string, error) {
		if len(lines) < 3 {
			return []string{}, nil
		}
		result := make([]string, 0, len(lines)-2)
		for _, l := range lines[2:] {
			f := strings.Split(l, "|")
			if len(f) != 5 {
				return nil, fmt.Errorf("expected 5 fields in line %q", l)
			}
			name := strings.TrimSpace(f[2])
			result = append(result, name)
		}
		return result, nil
	}

	return cmd, []string{}, parse
}

func (c ZypperCommandFactory) NewUpdateIndexCmd() (cmd, capabilities []string) {
	return []string{}, []string{}
}

func (c ZypperCommandFactory) NewUpgradeCmd() (cmd, capabilities []string) {
	cmd = []string{"zypper", "--non-interactive", "--quiet", "patch"}
	return cmd, []string{}
}

func (c ZypperCommandFactory) PackageManager() Manager {
	return Zypper
}
