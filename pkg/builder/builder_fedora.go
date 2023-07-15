// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"fmt"

	"github.com/ok-ryoko/turret/pkg/linux"

	"github.com/containers/buildah"
)

type FedoraTurretBuilder struct {
	TurretBuilder
}

func (b *FedoraTurretBuilder) CleanPackageCaches() error {
	cmd := []string{"dnf", "clean", "all"}
	ro := b.defaultRunOptions()
	if err := b.run(cmd, ro); err != nil {
		return fmt.Errorf("removing cached dnf data: %w", err)
	}
	return nil
}

func (b *FedoraTurretBuilder) Distro() linux.LinuxDistro {
	return linux.Fedora
}

func (b *FedoraTurretBuilder) InstallPackages(packages []string) error {
	if len(packages) == 0 {
		return nil
	}

	cmd := []string{"dnf", "--assumeyes", "--setopt=install_weak_deps=False", "install"}
	cmd = append(cmd, packages...)

	ro := b.defaultRunOptions()
	ro.AddCapabilities = []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_SETFCAP",
	}
	ro.ConfigureNetwork = buildah.NetworkEnabled

	if err := b.run(cmd, ro); err != nil {
		return fmt.Errorf("installing dnf packages: %w", err)
	}

	return nil
}

func (b *FedoraTurretBuilder) UpgradePackages() error {
	cmd := []string{"dnf", "--assumeyes", "--refresh", "upgrade"}

	ro := b.defaultRunOptions()
	ro.AddCapabilities = []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_SETFCAP",
	}
	ro.ConfigureNetwork = buildah.NetworkEnabled

	if err := b.run(cmd, ro); err != nil {
		return fmt.Errorf("upgrading dnf packages: %w", err)
	}

	return nil
}
