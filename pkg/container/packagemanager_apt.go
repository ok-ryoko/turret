// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"fmt"

	"github.com/containers/buildah"
)

type APTPackageManager struct {
	PackageManager
}

func (pm *APTPackageManager) Install(c *Container, packages []string) error {
	cmd, capabilities := pm.NewUpdateIndexCmd()
	ro := c.defaultRunOptions()
	ro.AddCapabilities = capabilities
	ro.ConfigureNetwork = buildah.NetworkEnabled
	if err := c.run(cmd, ro); err != nil {
		return fmt.Errorf(
			"updating %s package index: %w",
			pm.PackageManager.PackageManager().String(),
			err,
		)
	}

	cmd, capabilities = pm.NewInstallCmd(packages)
	ro = c.defaultRunOptions()
	ro.AddCapabilities = capabilities
	ro.ConfigureNetwork = buildah.NetworkEnabled
	if err := c.run(cmd, ro); err != nil {
		return fmt.Errorf(
			"installing %s packages: %w",
			pm.PackageManager.PackageManager().String(),
			err,
		)
	}
	return nil
}

func (pm *APTPackageManager) Upgrade(c *Container) error {
	cmd, capabilities := pm.NewUpdateIndexCmd()
	ro := c.defaultRunOptions()
	ro.AddCapabilities = capabilities
	ro.ConfigureNetwork = buildah.NetworkEnabled
	if err := c.run(cmd, ro); err != nil {
		return fmt.Errorf(
			"updating %s package index: %w",
			pm.PackageManager.PackageManager().String(),
			err,
		)
	}

	cmd, capabilities = pm.NewUpgradeCmd()
	ro = c.defaultRunOptions()
	ro.AddCapabilities = capabilities
	ro.ConfigureNetwork = buildah.NetworkEnabled
	if err := c.run(cmd, ro); err != nil {
		return fmt.Errorf(
			"upgrading pre-installed %s packages: %w",
			pm.PackageManager.PackageManager().String(),
			err,
		)
	}
	return nil
}
