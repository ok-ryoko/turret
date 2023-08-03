// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"fmt"

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
	List(c *Container) error

	// Upgrade upgrades the packages in the working container.
	Upgrade(c *Container) error
}

// PackageManager provides a high-level front end for Buildah for managing
// packages in a Linux builder container.
type PackageManager struct {
	pckg.CommandFactory
}

// CleanCaches cleans the package caches in the working container.
func (pm *PackageManager) CleanCaches(c *Container) error {
	cmd, capabilities := pm.NewCleanCacheCmd()
	ro := c.DefaultRunOptions()
	ro.AddCapabilities = capabilities
	errMsg := fmt.Sprintf("cleaning %s package caches", pm.PackageManager().String())
	if err := c.runWithLogging(cmd, ro, errMsg); err != nil {
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
	errMsg := fmt.Sprintf("installing %s packages", pm.PackageManager().String())
	if err := c.runWithLogging(cmd, ro, errMsg); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

// List lists the packages installed in the working container.
func (pm *PackageManager) List(c *Container) error {
	cmd, capabilities := pm.NewListInstalledPackagesCmd()
	ro := c.DefaultRunOptions()
	ro.AddCapabilities = capabilities
	errMsg := fmt.Sprintf("listing installed %s packages", pm.PackageManager().String())
	if err := c.runWithLogging(cmd, ro, errMsg); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

// Upgrade upgrades the packages in the working container.
func (pm *PackageManager) Upgrade(c *Container) error {
	cmd, capabilities := pm.NewUpgradeCmd()
	ro := c.DefaultRunOptions()
	ro.AddCapabilities = capabilities
	ro.ConfigureNetwork = buildah.NetworkEnabled
	errMsg := fmt.Sprintf("upgrading pre-installed %s packages", pm.PackageManager().String())
	if err := c.runWithLogging(cmd, ro, errMsg); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

// NewPackageManager creates a new PackageManager for a particular package
// manager.
func NewPackageManager(pm pckg.Manager) (PackageManagerInterface, error) {
	cf, err := pckg.NewCommandFactory(pm)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	var tpm PackageManagerInterface
	switch pm {
	case pckg.APT:
		tpm = &APTPackageManager{PackageManager{cf}}
	case
		pckg.APK,
		pckg.DNF,
		pckg.Pacman,
		pckg.XBPS,
		pckg.Zypper:
		tpm = &PackageManager{cf}
	default:
		return nil, fmt.Errorf("unrecognized package manager")
	}
	return tpm, nil
}
