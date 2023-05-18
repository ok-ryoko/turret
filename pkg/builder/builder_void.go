package builder

import (
	"fmt"

	"github.com/containers/buildah"
)

type VoidTurretBuilder struct {
	TurretBuilder
}

func (b *VoidTurretBuilder) CleanPackageCaches() error {
	cmd := []string{"xbps-remove", "--clean-cache", "--yes"}
	ro := b.defaultRunOptions()
	if err := b.run(cmd, ro); err != nil {
		return fmt.Errorf("cleaning xbps package cache: %w", err)
	}
	return nil
}

func (b *VoidTurretBuilder) Distro() GNULinuxDistro {
	return Void
}

func (b *VoidTurretBuilder) InstallPackages(packages []string) error {
	if len(packages) == 0 {
		return nil
	}

	cmd := []string{"xbps-install", "--yes"}
	cmd = append(cmd, packages...)

	ro := b.defaultRunOptions()
	ro.AddCapabilities = []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_SETFCAP",
	}
	ro.ConfigureNetwork = buildah.NetworkEnabled

	if err := b.run(cmd, ro); err != nil {
		return fmt.Errorf("installing xbps packages: %w", err)
	}

	return nil
}

func (b *VoidTurretBuilder) UpgradePackages() error {
	cmd := []string{"xbps-install", "--sync", "--update", "--yes"}

	ro := b.defaultRunOptions()
	ro.AddCapabilities = []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_SETFCAP",
	}
	ro.ConfigureNetwork = buildah.NetworkEnabled

	if err := b.run(cmd, ro); err != nil {
		return fmt.Errorf("updating xbps packages: %w", err)
	}

	return nil
}
