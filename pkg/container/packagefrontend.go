// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"fmt"
	"strings"

	"github.com/containers/buildah"
	"github.com/ok-ryoko/turret/pkg/linux/pckg"
)

// PackageFrontendInterface is the interface implemented by a PackageFrontend
// for a particular package manager.
type PackageFrontendInterface interface {
	// CleanCaches cleans the package caches in the working container.
	CleanCaches(c *Container) error

	// Install installs one or more packages to the working container.
	Install(c *Container, packages []string) error

	// List lists the packages installed in the working container.
	List(c *Container) ([]string, error)

	// Upgrade upgrades the packages in the working container.
	Upgrade(c *Container) error
}

// PackageFrontend provides a high-level frontend for Buildah for managing
// packages in a Linux builder container.
type PackageFrontend struct {
	pckg.CommandFactory
}

// CleanCaches cleans the package caches in the working container.
func (f *PackageFrontend) CleanCaches(c *Container) error {
	cmd, capabilities := f.NewCleanCacheCmd()
	ro := c.DefaultRunOptions()
	ro.AddCapabilities = capabilities
	errContext := fmt.Sprintf("cleaning %s package caches", f.Backend())
	if err := c.runWithLogging(cmd, ro, errContext); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

// Install installs one or more packages to the working container.
func (f *PackageFrontend) Install(c *Container, packages []string) error {
	cmd, capabilities := f.NewInstallCmd(packages)
	ro := c.DefaultRunOptions()
	ro.AddCapabilities = capabilities
	ro.ConfigureNetwork = buildah.NetworkEnabled
	errContext := fmt.Sprintf("installing %s packages", f.Backend())
	if err := c.runWithLogging(cmd, ro, errContext); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

// List lists the packages installed in the working container.
func (f *PackageFrontend) List(c *Container) ([]string, error) {
	cmd, capabilities, parse := f.NewListInstalledPackagesCmd()

	ro := c.DefaultRunOptions()
	ro.AddCapabilities = capabilities

	outText, errText, err := c.Run(cmd, ro)
	errContext := fmt.Sprintf("listing installed %s packages", f.Backend())
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
func (f *PackageFrontend) Upgrade(c *Container) error {
	cmd, capabilities := f.NewUpgradeCmd()
	ro := c.DefaultRunOptions()
	ro.AddCapabilities = capabilities
	ro.ConfigureNetwork = buildah.NetworkEnabled
	errContext := fmt.Sprintf("upgrading pre-installed %s packages", f.Backend())
	if err := c.runWithLogging(cmd, ro, errContext); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

// NewPackageFrontend creates a frontend for a particular package manager.
func NewPackageFrontend(backend pckg.Backend) (PackageFrontendInterface, error) {
	factory, err := pckg.NewCommandFactory(backend)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	var result PackageFrontendInterface
	switch backend {
	case pckg.APT:
		result = &APTPackageFrontend{PackageFrontend{factory}}
	case
		pckg.APK,
		pckg.DNF,
		pckg.Pacman,
		pckg.XBPS,
		pckg.Zypper:
		result = &PackageFrontend{factory}
	default:
		return nil, fmt.Errorf("unrecognized package manager %v", backend)
	}
	return result, nil
}
