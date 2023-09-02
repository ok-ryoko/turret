// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package pckg

import "fmt"

// CommandFactory provides a layer of abstraction over package management
// operations.
type CommandFactory interface {
	// NewCleanCacheCmd returns (1) a command that cleans the package cache and
	// (2) the Linux capabilities needed by that command.
	NewCleanCacheCmd() (cmd, capabilities []string)

	// NewInstallCmd returns (1) a command that installs one or more packages
	// and (2) the Linux capabilities needed by that command.
	NewInstallCmd(packages []string) (cmd, capabilities []string)

	// NewListInstalledPackagesCmd returns:
	//
	//   (1) a command that lists the installed packages;
	//   (2) the Linux capabilities needed by that command, and
	//   (3) a function to parse the package names from the command's output.
	NewListInstalledPackagesCmd() (cmd, capabilities []string, parse func([]string) ([]string, error))

	// NewUpdateIndexCmd returns (1) a command that updates the package index
	// and (2) the Linux capabilities needed by that command.
	NewUpdateIndexCmd() (cmd, capabilities []string)

	// NewUpgradeCmd returns a command that upgrades the pre-installed packages
	// and (2) the Linux capabilities needed by that command.
	NewUpgradeCmd() (cmd, capabilities []string)

	// Backend returns a constant representing the package manager for which
	// this factory makes commands.
	Backend() Backend
}

// NewCommandFactory creates an object that manufactures package management
// commands for execution in a shell.
func NewCommandFactory(b Backend) (CommandFactory, error) {
	var factory CommandFactory
	switch b {
	case APK:
		factory = &APKCommandFactory{}
	case APT:
		factory = &APTCommandFactory{}
	case DNF:
		factory = &DNFCommandFactory{}
	case Pacman:
		factory = &PacmanCommandFactory{}
	case XBPS:
		factory = &XBPSCommandFactory{}
	case Zypper:
		factory = &ZypperCommandFactory{}
	default:
		return nil, fmt.Errorf("unrecognized package manager %v", b)
	}
	return factory, nil
}
