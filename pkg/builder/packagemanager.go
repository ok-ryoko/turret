// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"fmt"

	"github.com/containers/buildah"
	"github.com/ok-ryoko/turret/pkg/linux/packagemanager"
)

// TurretPackageManagerInterface is the interface implemented by a
// TurretPackageManager for a particular package manager.
type TurretPackageManagerInterface interface {
	// CleanCaches cleans the package caches in the working container.
	CleanCaches(b *TurretBuilder) error

	// Install installs one or more packages to the working container.
	Install(b *TurretBuilder, packages []string) error

	// List lists the packages installed in the working container.
	List(b *TurretBuilder) error

	// Upgrade upgrades the packages in the working container.
	Upgrade(b *TurretBuilder) error
}

// TurretPackageManager provides a high-level front end for Buildah for
// managing packages in a Linux builder container.
type TurretPackageManager struct {
	packagemanager.CommandFactory
}

// CleanCaches cleans the package caches in the working container.
func (pm *TurretPackageManager) CleanCaches(b *TurretBuilder) error {
	cmd, capabilities := pm.NewCleanCacheCmd()
	ro := b.defaultRunOptions()
	ro.AddCapabilities = capabilities
	if err := b.run(cmd, ro); err != nil {
		return fmt.Errorf(
			"cleaning %s package cache: %w",
			pm.PackageManager().String(),
			err,
		)
	}
	return nil
}

// Install installs one or more packages to the working container.
func (pm *TurretPackageManager) Install(b *TurretBuilder, packages []string) error {
	cmd, capabilities := pm.NewInstallCmd(packages)
	ro := b.defaultRunOptions()
	ro.AddCapabilities = capabilities
	ro.ConfigureNetwork = buildah.NetworkEnabled
	if err := b.run(cmd, ro); err != nil {
		return fmt.Errorf(
			"installing %s packages: %w",
			pm.PackageManager().String(),
			err,
		)
	}
	return nil
}

// List lists the packages installed in the working container.
func (pm *TurretPackageManager) List(b *TurretBuilder) error {
	cmd, capabilities := pm.NewListInstalledPackagesCmd()
	ro := b.defaultRunOptions()
	ro.AddCapabilities = capabilities
	if err := b.run(cmd, ro); err != nil {
		return fmt.Errorf(
			"listing installed %s packages: %w",
			pm.PackageManager().String(),
			err,
		)
	}
	return nil
}

// Upgrade upgrades the packages in the working container.
func (pm *TurretPackageManager) Upgrade(b *TurretBuilder) error {
	cmd, capabilities := pm.NewUpgradeCmd()
	ro := b.defaultRunOptions()
	ro.AddCapabilities = capabilities
	ro.ConfigureNetwork = buildah.NetworkEnabled
	if err := b.run(cmd, ro); err != nil {
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
func NewPackageManager(pm packagemanager.PackageManager) (TurretPackageManagerInterface, error) {
	cmdFactory, err := packagemanager.New(pm)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	var tpm TurretPackageManagerInterface
	switch pm {
	case packagemanager.APT:
		tpm = &APTTurretPackageManager{TurretPackageManager{cmdFactory}}
	case
		packagemanager.APK,
		packagemanager.DNF,
		packagemanager.Pacman,
		packagemanager.XBPS,
		packagemanager.Zypper:
		tpm = &TurretPackageManager{cmdFactory}
	default:
		return nil, fmt.Errorf("unrecognized package manager")
	}
	return tpm, nil
}
