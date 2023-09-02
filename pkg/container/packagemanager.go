// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"fmt"
	"strings"

	"github.com/containers/buildah"
	"github.com/ok-ryoko/turret/pkg/linux/pckg"
)

// PackageManagerInterface is the interface implemented by a PackageManager
// for a particular package manager.
type PackageManagerInterface interface {
	// CleanCaches cleans the package caches in the working container.
	CleanCaches(c *Container) error

	// Install installs one or more packages to the working container.
	Install(c *Container, packages []string) error

	// List lists the packages installed in the working container.
	List(c *Container) ([]string, error)

	// Upgrade upgrades the packages in the working container.
	Upgrade(c *Container) error
}

// PackageManager provides a high-level frontend for Buildah for managing
// packages in a Linux builder container.
type PackageManager struct {
	pckg.CommandFactory
}

// CleanCaches cleans the package caches in the working container.
func (pm *PackageManager) CleanCaches(c *Container) error {
	cmd, capabilities := pm.NewCleanCacheCmd()
	ro := c.DefaultRunOptions()
	ro.AddCapabilities = capabilities
	errContext := fmt.Sprintf("cleaning %s package caches", pm.PackageManager())
	if err := c.runWithLogging(cmd, ro, errContext); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

// Install installs one or more packages to the working container.
func (pm *PackageManager) Install(c *Container, packages []string) error {
	cmd, capabilities := pm.NewInstallCmd(packages)
	ro := c.DefaultRunOptions()
	ro.AddCapabilities = capabilities
	ro.ConfigureNetwork = buildah.NetworkEnabled
	errContext := fmt.Sprintf("installing %s packages", pm.PackageManager())
	if err := c.runWithLogging(cmd, ro, errContext); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

// List lists the packages installed in the working container.
func (pm *PackageManager) List(c *Container) ([]string, error) {
	cmd, capabilities, parse := pm.NewListInstalledPackagesCmd()

	ro := c.DefaultRunOptions()
	ro.AddCapabilities = capabilities

	outText, errText, err := c.Run(cmd, ro)
	errContext := fmt.Sprintf("listing installed %s packages", pm.PackageManager())
	if err != nil {
		if errText != "" {
			errContext = fmt.Sprintf("%s (%q)", errContext, errText)
		}
		return nil, fmt.Errorf("%s: %w", errContext, err)
	}

	lines := strings.Split(strings.ReplaceAll(strings.TrimSpace(outText), "\r\n", "\n"), "\n")
	packages, err := parse(lines)
	if err != nil {
		return nil, fmt.Errorf("parsing installed packages: %w", err)
	}

	return packages, nil
}

// Upgrade upgrades the packages in the working container.
func (pm *PackageManager) Upgrade(c *Container) error {
	cmd, capabilities := pm.NewUpgradeCmd()
	ro := c.DefaultRunOptions()
	ro.AddCapabilities = capabilities
	ro.ConfigureNetwork = buildah.NetworkEnabled
	errContext := fmt.Sprintf("upgrading pre-installed %s packages", pm.PackageManager())
	if err := c.runWithLogging(cmd, ro, errContext); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

// NewPackageManager creates a frontend for a particular package manager.
func NewPackageManager(manager pckg.Manager) (PackageManagerInterface, error) {
	factory, err := pckg.NewCommandFactory(manager)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	var result PackageManagerInterface
	switch manager {
	case pckg.APT:
		result = &APTPackageManager{PackageManager{factory}}
	case
		pckg.APK,
		pckg.DNF,
		pckg.Pacman,
		pckg.XBPS,
		pckg.Zypper:
		result = &PackageManager{factory}
	default:
		return nil, fmt.Errorf("unrecognized package manager %v", manager)
	}
	return result, nil
}
