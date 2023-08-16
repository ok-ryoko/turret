// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package pckg

import (
	"fmt"
	"strings"
)

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

func (c XBPSCommandFactory) NewListInstalledPackagesCmd() (
	cmd []string,
	capabilities []string,
	parse func([]string) ([]string, error),
) {
	cmd = []string{"xbps-query", "--list-pkgs"}
	capabilities = []string{}

	// expected line format: status name-version_revision description
	parse = func(lines []string) ([]string, error) {
		result := make([]string, 0, len(lines))
		for _, l := range lines {
			f := strings.Fields(l)
			if len(f) < 3 {
				return nil, fmt.Errorf("expected at least 3 fields in line %q", l)
			}
			i := strings.LastIndex(f[1], "-")
			if i == -1 {
				return nil, fmt.Errorf("expected format 'name-version_revision' for field %q", f[1])
			}
			name := f[1][:i]
			result = append(result, name)
		}
		return result, nil
	}

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
