// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"fmt"

	"github.com/ok-ryoko/turret/pkg/linux"

	"github.com/containers/buildah"
)

type ArchTurretBuilder struct {
	TurretBuilder
}

func (b *ArchTurretBuilder) CleanPackageCaches() error {
	cmd := []string{"pacman", "--sync", "--clean", "--clean", "--noconfirm"}
	ro := b.defaultRunOptions()
	if err := b.run(cmd, ro); err != nil {
		return fmt.Errorf("cleaning pacman package cache: %w", err)
	}
	return nil
}

func (b *ArchTurretBuilder) Distro() linux.LinuxDistro {
	return linux.Arch
}

func (b *ArchTurretBuilder) InstallPackages(packages []string) error {
	if len(packages) == 0 {
		return nil
	}

	cmd := []string{"pacman", "--sync", "--noconfirm", "--noprogressbar"}
	cmd = append(cmd, packages...)

	ro := b.defaultRunOptions()
	ro.AddCapabilities = []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_FOWNER",
		"CAP_SYS_CHROOT",
	}
	ro.ConfigureNetwork = buildah.NetworkEnabled

	if err := b.run(cmd, ro); err != nil {
		return fmt.Errorf("installing pacman packages: %w", err)
	}

	return nil
}

func (b *ArchTurretBuilder) UpgradePackages() error {
	cmd := []string{
		"pacman",
		"--sync",
		"--sysupgrade",
		"--refresh",
		"--noconfirm",
		"--noprogressbar",
	}

	ro := b.defaultRunOptions()
	ro.AddCapabilities = []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_FOWNER",
		"CAP_SYS_CHROOT",
	}
	ro.ConfigureNetwork = buildah.NetworkEnabled

	if err := b.run(cmd, ro); err != nil {
		return fmt.Errorf("upgrading system: %w", err)
	}

	return nil
}
