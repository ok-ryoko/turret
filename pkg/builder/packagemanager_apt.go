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

func (p *APTTurretPackageManager) Install(b *TurretBuilder, packages []string) error {
	cmd, capabilities := p.NewUpdateIndexCmd()
	ro := b.defaultRunOptions()
	ro.AddCapabilities = capabilities
	ro.ConfigureNetwork = buildah.NetworkEnabled
	if err := b.run(cmd, ro); err != nil {
		return fmt.Errorf(
			"updating %s package index: %w",
			p.PackageManager().String(),
			err,
		)
	}

	cmd, capabilities = p.NewInstallCmd(packages)
	ro = b.defaultRunOptions()
	ro.AddCapabilities = capabilities
	ro.ConfigureNetwork = buildah.NetworkEnabled
	if err := b.run(cmd, ro); err != nil {
		return fmt.Errorf(
			"installing %s packages: %w",
			p.PackageManager().String(),
			err,
		)
	}
	return nil
}

func (p *APTTurretPackageManager) Upgrade(b *TurretBuilder) error {
	cmd, capabilities := p.NewUpdateIndexCmd()
	ro := b.defaultRunOptions()
	ro.AddCapabilities = capabilities
	ro.ConfigureNetwork = buildah.NetworkEnabled
	if err := b.run(cmd, ro); err != nil {
		return fmt.Errorf(
			"updating %s package index: %w",
			p.PackageManager().String(),
			err,
		)
	}

	cmd, capabilities = p.NewUpgradeCmd()
	ro = b.defaultRunOptions()
	ro.AddCapabilities = capabilities
	ro.ConfigureNetwork = buildah.NetworkEnabled
	if err := b.run(cmd, ro); err != nil {
		return fmt.Errorf(
			"upgrading pre-installed %s packages: %w",
			p.PackageManager().String(),
			err,
		)
	}
	return nil
}
