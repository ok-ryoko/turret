package builder

import (
	"fmt"

	"github.com/containers/buildah"
)

type DebianTurretBuilder struct {
	TurretBuilder
}

func (b *DebianTurretBuilder) CleanPackageCaches() error {
	cmd := []string{"apt", "clean"}
	ro := b.defaultRunOptions()
	if err := b.run(cmd, ro); err != nil {
		return fmt.Errorf("cleaning apt package cache: %w", err)
	}
	return nil
}

func (b *DebianTurretBuilder) Distro() GNULinuxDistro {
	return Debian
}

func (b *DebianTurretBuilder) InstallPackages(packages []string) error {
	if len(packages) == 0 {
		return nil
	}

	updateCmd := []string{"apt", "update"}

	ro := b.defaultRunOptions()
	ro.ConfigureNetwork = buildah.NetworkEnabled

	if err := b.run(updateCmd, ro); err != nil {
		return fmt.Errorf("updating apt package index: %w", err)
	}

	installCmd := []string{"apt", "--yes", "install"}
	installCmd = append(installCmd, packages...)

	ro.AddCapabilities = []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_SETFCAP",
	}

	if err := b.run(installCmd, ro); err != nil {
		return fmt.Errorf("installing apt packages: %w", err)
	}

	return nil
}

func (b *DebianTurretBuilder) UpgradePackages() error {
	updateCmd := []string{"apt", "--quiet", "update"}

	ro := b.defaultRunOptions()
	ro.ConfigureNetwork = buildah.NetworkEnabled

	if err := b.run(updateCmd, ro); err != nil {
		return fmt.Errorf("updating apt package index: %w", err)
	}

	upgradeCmd := []string{"apt", "--yes", "upgrade"}

	ro.AddCapabilities = []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_SETFCAP",
	}

	if err := b.run(upgradeCmd, ro); err != nil {
		return fmt.Errorf("upgrading apt packages: %w", err)
	}

	return nil
}
