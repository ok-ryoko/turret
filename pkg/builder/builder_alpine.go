// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"fmt"

	"github.com/containers/buildah"
)

type AlpineTurretBuilder struct {
	TurretBuilder
}

func (b *AlpineTurretBuilder) CleanPackageCaches() error {
	return nil
}

func (b *AlpineTurretBuilder) CreateUser(name string, distro GNULinuxDistro, options CreateUserOptions) error {
	if name == "" {
		return fmt.Errorf("blank user name")
	}

	shell, err := b.resolveExecutable(options.LoginShell, distro)
	if err != nil {
		return fmt.Errorf("resolving login shell: %w", err)
	}
	options.LoginShell = shell

	adduserCmd := []string{
		"adduser",
		"-D",
		"-G", "users",
		"-u", fmt.Sprintf("%d", options.UID),
		"-s", options.LoginShell,
	}

	if options.Comment != "" {
		adduserCmd = append(adduserCmd, "-g", options.Comment)
	}

	adduserCmd = append(adduserCmd, name)

	ro := b.defaultRunOptions()
	ro.AddCapabilities = []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_FOWNER",
	}

	if err := b.run(adduserCmd, ro); err != nil {
		return fmt.Errorf("%w", err)
	}

	if len(options.Groups) > 0 {
		ago := b.defaultRunOptions()
		for _, g := range options.Groups {
			addgroupCmd := []string{"addgroup", name, g}
			if err := b.run(addgroupCmd, ago); err != nil {
				return fmt.Errorf("%w", err)
			}
		}
	}

	return nil
}

func (b *AlpineTurretBuilder) Distro() GNULinuxDistro {
	return Alpine
}

func (b *AlpineTurretBuilder) InstallPackages(packages []string) error {
	if len(packages) == 0 {
		return nil
	}

	cmd := []string{"apk", "--no-cache", "--no-progress", "add"}
	cmd = append(cmd, packages...)

	ro := b.defaultRunOptions()
	ro.AddCapabilities = []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_SETFCAP",
	}
	ro.ConfigureNetwork = buildah.NetworkEnabled

	if err := b.run(cmd, ro); err != nil {
		return fmt.Errorf("adding apk packages: %w", err)
	}

	return nil
}

func (b *AlpineTurretBuilder) UpgradePackages() error {
	cmd := []string{"apk", "--no-cache", "--no-progress", "upgrade"}

	ro := b.defaultRunOptions()
	ro.AddCapabilities = []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_SETFCAP",
	}
	ro.ConfigureNetwork = buildah.NetworkEnabled

	if err := b.run(cmd, ro); err != nil {
		return fmt.Errorf("upgrading apk packages: %w", err)
	}

	return nil
}
