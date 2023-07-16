// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package packagemanager

import (
	"fmt"
	"strings"
)

// PackageManager is an identifier for a package manager for Linux-based
// distros; the zero value represents an unknown package manager
type PackageManager int

const (
	APK = 1 << iota
	APT
	DNF
	Pacman
	XBPS
	Zypper
)

func (p PackageManager) String() string {
	var s string
	switch p {
	case APK:
		s = "APK"
	case APT:
		s = "APT"
	case DNF:
		s = "DNF"
	case Pacman:
		s = "Pacman"
	case XBPS:
		s = "XBPS"
	case Zypper:
		s = "Zypper"
	default:
		s = "unknown"
	}
	return s
}

// RePackageName returns a regular expression to match valid package names for
// the package manager's ecosystem
func (p PackageManager) RePackageName() string {
	var r string
	switch p {
	case APK:
		r = `^[0-9a-z][+-\._0-9a-z]*[0-9a-z]$`
	case APT:
		r = `^[0-9a-z][+-\.0-9a-z]*[0-9a-z]$`
	case DNF, XBPS, Zypper:
		r = `^[0-9A-Za-z][+-\._0-9A-Za-z]*[0-9A-Za-z]$`
	case Pacman:
		r = `^[0-9a-z][+-\._0-9a-z]*[0-9a-z]$`
	default:
		r = ""
	}
	return r
}

// PackageManagerWrapper wraps PackageManager to facilitate its parsing
type PackageManagerWrapper struct {
	PackageManager
}

// UnmarshalText decodes the package manager from a string
func (p *PackageManagerWrapper) UnmarshalText(text []byte) error {
	var err error
	p.PackageManager, err = parsePackageManagerString(string(text))
	return err
}

func parsePackageManagerString(s string) (PackageManager, error) {
	var p PackageManager
	switch strings.ToLower(s) {
	case "apk":
		p = APK
	case "apt":
		p = APT
	case "dnf":
		p = DNF
	case "pacman":
		p = Pacman
	case "xbps":
		p = XBPS
	case "zypper":
		p = Zypper
	default:
		return 0, fmt.Errorf("unsupported package manager: %s", s)
	}
	return p, nil
}

// PackageManagerCommandFactory provides a simple layer of abstraction over
// common package manager commands
type CommandFactory interface {
	// NewCleanCacheCmd returns (1) a command that cleans the package cache and
	// (2) the Linux capabilities needed by that command
	NewCleanCacheCmd() (cmd, capabilities []string)

	// NewInstallCmd returns (1) a command that installs one or more packages
	// and (2) the Linux capabilities needed by that command
	NewInstallCmd(packages []string) (cmd, capabilities []string)

	// NewListInstalledPackagesCmd returns (1) a command that lists the
	// installed packages and (2) the Linux capabilities needed by that command
	NewListInstalledPackagesCmd() (cmd, capabilities []string)

	// NewUpdateIndexCmd returns (1) a command that updates the package index
	// and (2) the Linux capabilities needed by that command
	NewUpdateIndexCmd() (cmd, capabilities []string)

	// NewUpgradeCmd returns a command that upgrades the pre-installed packages
	// and (2) the Linux capabilities needed by that command
	NewUpgradeCmd() (cmd, capabilities []string)

	// PackageManager returns a constant representing the package manager for
	// which this factory makes commands
	PackageManager() PackageManager
}

func New(p PackageManager) (CommandFactory, error) {
	var result CommandFactory
	switch p {
	case APK:
		result = &APKPackageManager{}
	case APT:
		result = &APTPackageManager{}
	case DNF:
		result = &DNFPackageManager{}
	case Pacman:
		result = &PacmanPackageManager{}
	case XBPS:
		result = &XBPSPackageManager{}
	case Zypper:
		result = &ZypperPackageManager{}
	default:
		return nil, fmt.Errorf("unrecognized package manager")
	}
	return result, nil
}
