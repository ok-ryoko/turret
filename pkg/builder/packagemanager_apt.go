// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"fmt"

	"github.com/containers/buildah"
)

type APTTurretPackageManager struct {
	TurretPackageManager
}

func (pm *APTTurretPackageManager) Install(b *TurretBuilder, packages []string) error {
	cmd, capabilities := pm.NewUpdateIndexCmd()
	ro := b.defaultRunOptions()
	ro.AddCapabilities = capabilities
	ro.ConfigureNetwork = buildah.NetworkEnabled
	if err := b.run(cmd, ro); err != nil {
		return fmt.Errorf(
			"updating %s package index: %w",
			pm.PackageManager().String(),
			err,
		)
	}

	cmd, capabilities = pm.NewInstallCmd(packages)
	ro = b.defaultRunOptions()
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

func (pm *APTTurretPackageManager) Upgrade(b *TurretBuilder) error {
	cmd, capabilities := pm.NewUpdateIndexCmd()
	ro := b.defaultRunOptions()
	ro.AddCapabilities = capabilities
	ro.ConfigureNetwork = buildah.NetworkEnabled
	if err := b.run(cmd, ro); err != nil {
		return fmt.Errorf(
			"updating %s package index: %w",
			pm.PackageManager().String(),
			err,
		)
	}

	cmd, capabilities = pm.NewUpgradeCmd()
	ro = b.defaultRunOptions()
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
