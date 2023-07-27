// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"fmt"

	"github.com/containers/buildah"
	"github.com/ok-ryoko/turret/pkg/linux/pckg"
)

// TurretPackageManagerInterface is the interface implemented by a
// TurretPackageManager for a particular package manager.
type TurretPackageManagerInterface interface {
	// CleanCaches cleans the package caches in the working container.
	CleanCaches(c *TurretContainer) error

	// Install installs one or more packages to the working container.
	Install(c *TurretContainer, packages []string) error

	// List lists the packages installed in the working container.
	List(c *TurretContainer) error

	// Upgrade upgrades the packages in the working container.
	Upgrade(c *TurretContainer) error
}

// TurretPackageManager provides a high-level front end for Buildah for
// managing packages in a Linux builder container.
type TurretPackageManager struct {
	pckg.CommandFactory
}

// CleanCaches cleans the package caches in the working container.
func (pm *TurretPackageManager) CleanCaches(c *TurretContainer) error {
	cmd, capabilities := pm.NewCleanCacheCmd()
	ro := c.defaultRunOptions()
	ro.AddCapabilities = capabilities
	if err := c.run(cmd, ro); err != nil {
		return fmt.Errorf(
			"cleaning %s package cache: %w",
			pm.PackageManager().String(),
			err,
		)
	}
	return nil
}

// Install installs one or more packages to the working container.
func (pm *TurretPackageManager) Install(c *TurretContainer, packages []string) error {
	cmd, capabilities := pm.NewInstallCmd(packages)
	ro := c.defaultRunOptions()
	ro.AddCapabilities = capabilities
	ro.ConfigureNetwork = buildah.NetworkEnabled
	if err := c.run(cmd, ro); err != nil {
		return fmt.Errorf(
			"installing %s packages: %w",
			pm.PackageManager().String(),
			err,
		)
	}
	return nil
}

// List lists the packages installed in the working container.
func (pm *TurretPackageManager) List(c *TurretContainer) error {
	cmd, capabilities := pm.NewListInstalledPackagesCmd()
	ro := c.defaultRunOptions()
	ro.AddCapabilities = capabilities
	if err := c.run(cmd, ro); err != nil {
		return fmt.Errorf(
			"listing installed %s packages: %w",
			pm.PackageManager().String(),
			err,
		)
	}
	return nil
}

// Upgrade upgrades the packages in the working container.
func (pm *TurretPackageManager) Upgrade(c *TurretContainer) error {
	cmd, capabilities := pm.NewUpgradeCmd()
	ro := c.defaultRunOptions()
	ro.AddCapabilities = capabilities
	ro.ConfigureNetwork = buildah.NetworkEnabled
	if err := c.run(cmd, ro); err != nil {
		return fmt.Errorf(
			"upgrading pre-installed %s packages: %w",
			pm.PackageManager().String(),
			err,
		)
	}
	return nil
}

// NewPackageManager creates a new TurretPackageManager for a particular
// package manager.
func NewPackageManager(pm pckg.Manager) (TurretPackageManagerInterface, error) {
	cf, err := pckg.NewCommandFactory(pm)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	var tpm TurretPackageManagerInterface
	switch pm {
	case pckg.APT:
		tpm = &APTTurretPackageManager{TurretPackageManager{cf}}
	case
		pckg.APK,
		pckg.DNF,
		pckg.Pacman,
		pckg.XBPS,
		pckg.Zypper:
		tpm = &TurretPackageManager{cf}
	default:
		return nil, fmt.Errorf("unrecognized package manager")
	}
	return tpm, nil
}
