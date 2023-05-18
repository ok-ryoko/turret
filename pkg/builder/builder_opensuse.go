// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"fmt"

	"github.com/containers/buildah"
)

type OpenSUSETurretBuilder struct {
	TurretBuilder
}

func (b *OpenSUSETurretBuilder) CleanPackageCaches() error {
	cmd := []string{"zypper", "--non-interactive", "clean", "--all"}
	ro := b.defaultRunOptions()
	if err := b.run(cmd, ro); err != nil {
		return fmt.Errorf("cleaning local zypper caches: %w", err)
	}
	return nil
}

func (b *OpenSUSETurretBuilder) Distro() GNULinuxDistro {
	return OpenSUSE
}

func (b *OpenSUSETurretBuilder) InstallPackages(packages []string) error {
	if len(packages) == 0 {
		return nil
	}

	cmd := []string{"zypper", "--non-interactive", "install", "--no-recommends"}
	cmd = append(cmd, packages...)

	ro := b.defaultRunOptions()
	ro.AddCapabilities = []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_SETFCAP",
	}
	ro.ConfigureNetwork = buildah.NetworkEnabled

	if err := b.run(cmd, ro); err != nil {
		return fmt.Errorf("installing zypper packages: %w", err)
	}

	return nil
}

func (b *OpenSUSETurretBuilder) UpgradePackages() error {
	cmd := []string{"zypper", "--non-interactive", "patch"}

	ro := b.defaultRunOptions()
	ro.AddCapabilities = []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_SETFCAP",
	}
	ro.ConfigureNetwork = buildah.NetworkEnabled

	if err := b.run(cmd, ro); err != nil {
		return fmt.Errorf("applying patches to system: %w", err)
	}

	return nil
}
