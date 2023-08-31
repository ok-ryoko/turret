// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package pckg

import (
	"fmt"
	"strings"
)

type APKCommandFactory struct{}

func (c APKCommandFactory) NewCleanCacheCmd() (cmd, capabilities []string) {
	return []string{}, []string{}
}

func (c APKCommandFactory) NewInstallCmd(packages []string) (cmd, capabilities []string) {
	cmd = []string{"apk", "--no-cache", "--no-progress", "--quiet", "add"}
	cmd = append(cmd, packages...)
	return cmd, []string{}
}

func (c APKCommandFactory) NewListInstalledPackagesCmd() (
	cmd []string,
	capabilities []string,
	parse func([]string) ([]string, error),
) {
	cmd = []string{
		"apk",
		"--no-interactive",
		"--no-network",
		"--quiet",
		"list",
		"--installed",
	}

	// expected line format: name-version-revision arch {origin} (licenses) [status]
	parse = func(lines []string) ([]string, error) {
		result := make([]string, 0, len(lines))
		for _, l := range lines {
			pkg, _, ok := strings.Cut(l, " ")
			if !ok {
				return nil, fmt.Errorf("expected space delimiter in line %q", l)
			}
			i := strings.LastIndex(pkg, "-")
			if i == -1 {
				return nil, fmt.Errorf("expected format 'name-version-revision' for field %q", pkg)
			}
			j := strings.LastIndex(pkg[:i], "-")
			if j == -1 {
				return nil, fmt.Errorf("expected format 'name-version-revision' for field %q", pkg)
			}
			name := pkg[:j]
			result = append(result, name)
		}
		return result, nil
	}

	return cmd, []string{}, parse
}

func (c APKCommandFactory) NewUpdateIndexCmd() (cmd, capabilities []string) {
	return []string{}, []string{}
}

func (c APKCommandFactory) NewUpgradeCmd() (cmd, capabilities []string) {
	cmd = []string{"apk", "--no-cache", "--no-progress", "--quiet", "upgrade"}
	return cmd, []string{}
}

func (c APKCommandFactory) PackageManager() Manager {
	return APK
}
