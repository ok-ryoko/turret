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
	pmDisplay := pm.PackageManager.PackageManager().String()

	{
		cmd, capabilities := pm.NewUpdateIndexCmd()
		ro := c.DefaultRunOptions()
		ro.AddCapabilities = capabilities
		ro.ConfigureNetwork = buildah.NetworkEnabled
		errMsg := fmt.Sprintf("updating %s package index", pmDisplay)
		if err := c.runWithLogging(cmd, ro, errMsg); err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	{
		cmd, capabilities := pm.NewInstallCmd(packages)
		ro := c.DefaultRunOptions()
		ro.AddCapabilities = capabilities
		ro.ConfigureNetwork = buildah.NetworkEnabled
		errMsg := fmt.Sprintf("installing %s packages", pmDisplay)
		if err := c.runWithLogging(cmd, ro, errMsg); err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	return nil
}

func (pm *APTPackageManager) Upgrade(c *Container) error {
	pmDisplay := pm.PackageManager.PackageManager().String()

	{
		cmd, capabilities := pm.NewUpdateIndexCmd()
		ro := c.DefaultRunOptions()
		ro.AddCapabilities = capabilities
		ro.ConfigureNetwork = buildah.NetworkEnabled
		errMsg := fmt.Sprintf("updating %s package index", pmDisplay)
		if err := c.runWithLogging(cmd, ro, errMsg); err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	{
		cmd, capabilities := pm.NewUpgradeCmd()
		ro := c.DefaultRunOptions()
		ro.AddCapabilities = capabilities
		ro.ConfigureNetwork = buildah.NetworkEnabled
		errMsg := fmt.Sprintf("upgrading pre-installed %s packages", pmDisplay)
		if err := c.runWithLogging(cmd, ro, errMsg); err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	return nil
}
