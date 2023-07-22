// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package pckg

import "fmt"

// CommandFactory provides a simple layer of abstraction over common package
// manager operations.
type CommandFactory interface {
	// NewCleanCacheCmd returns (1) a command that cleans the package cache and
	// (2) the Linux capabilities needed by that command.
	NewCleanCacheCmd() (cmd, capabilities []string)

	// NewInstallCmd returns (1) a command that installs one or more packages
	// and (2) the Linux capabilities needed by that command.
	NewInstallCmd(packages []string) (cmd, capabilities []string)

	// NewListInstalledPackagesCmd returns (1) a command that lists the
	// installed packages and (2) the Linux capabilities needed by that command.
	NewListInstalledPackagesCmd() (cmd, capabilities []string)

	// NewUpdateIndexCmd returns (1) a command that updates the package index
	// and (2) the Linux capabilities needed by that command.
	NewUpdateIndexCmd() (cmd, capabilities []string)

	// NewUpgradeCmd returns a command that upgrades the pre-installed packages
	// and (2) the Linux capabilities needed by that command.
	NewUpgradeCmd() (cmd, capabilities []string)

	// PackageManager returns a constant representing the package manager for
	// which this factory makes commands.
	PackageManager() Manager
}

// NewCommandFactory creates a new CommandFactory that manufactures package
// manager commands for execution in a shell.
func NewCommandFactory(m Manager) (CommandFactory, error) {
	var result CommandFactory
	switch m {
	case APK:
		result = &APKCommandFactory{}
	case APT:
		result = &APTCommandFactory{}
	case DNF:
		result = &DNFCommandFactory{}
	case Pacman:
		result = &PacmanCommandFactory{}
	case XBPS:
		result = &XBPSCommandFactory{}
	case Zypper:
		result = &ZypperCommandFactory{}
	default:
		return nil, fmt.Errorf("unrecognized package manager")
	}
	return result, nil
}
