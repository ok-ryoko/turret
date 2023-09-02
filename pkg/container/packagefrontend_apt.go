// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"fmt"

	"github.com/containers/buildah"
)

type APTPackageFrontend struct {
	PackageFrontend
}

func (f *APTPackageFrontend) Install(c *Container, packages []string) error {
	{
		cmd, capabilities := f.NewUpdateIndexCmd()
		ro := c.DefaultRunOptions()
		ro.AddCapabilities = capabilities
		ro.ConfigureNetwork = buildah.NetworkEnabled
		errContext := fmt.Sprintf("updating %s package index", f.Backend())
		if err := c.runWithLogging(cmd, ro, errContext); err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	{
		cmd, capabilities := f.NewInstallCmd(packages)
		ro := c.DefaultRunOptions()
		ro.AddCapabilities = capabilities
		ro.ConfigureNetwork = buildah.NetworkEnabled
		errContext := fmt.Sprintf("installing %s packages", f.Backend())
		if err := c.runWithLogging(cmd, ro, errContext); err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	return nil
}

func (f *APTPackageFrontend) Upgrade(c *Container) error {
	{
		cmd, capabilities := f.NewUpdateIndexCmd()
		ro := c.DefaultRunOptions()
		ro.AddCapabilities = capabilities
		ro.ConfigureNetwork = buildah.NetworkEnabled
		errContext := fmt.Sprintf("updating %s package index", f.Backend())
		if err := c.runWithLogging(cmd, ro, errContext); err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	{
		cmd, capabilities := f.NewUpgradeCmd()
		ro := c.DefaultRunOptions()
		ro.AddCapabilities = capabilities
		ro.ConfigureNetwork = buildah.NetworkEnabled
		errContext := fmt.Sprintf("upgrading pre-installed %s packages", f.Backend())
		if err := c.runWithLogging(cmd, ro, errContext); err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	return nil
}
