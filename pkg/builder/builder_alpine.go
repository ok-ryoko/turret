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

func (b *AlpineTurretBuilder) CreateUser(name string, distro LinuxDistro, options CreateUserOptions) error {
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

	// CAP_DAC_OVERRIDE and CAP_FSETID are elements of the useradd effective
	// capability set but are not needed for the operation to succeed
	//
	ro := b.defaultRunOptions()
	ro.AddCapabilities = []string{
		"CAP_CHOWN",
		//
		// Change owner of /home/user

		"CAP_FOWNER",
		//
		// Change mode and owner of /home/user as well as temporary files when
		// editing /etc/passwd, /etc/shadow and /etc/group
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

func (b *AlpineTurretBuilder) Distro() LinuxDistro {
	return Alpine
}

func (b *AlpineTurretBuilder) InstallPackages(packages []string) error {
	if len(packages) == 0 {
		return nil
	}

	cmd := []string{"apk", "--no-cache", "--no-progress", "add"}
	cmd = append(cmd, packages...)

	ro := b.defaultRunOptions()
	ro.ConfigureNetwork = buildah.NetworkEnabled

	if err := b.run(cmd, ro); err != nil {
		return fmt.Errorf("adding apk packages: %w", err)
	}

	return nil
}

func (b *AlpineTurretBuilder) UpgradePackages() error {
	cmd := []string{"apk", "--no-cache", "--no-progress", "upgrade"}

	ro := b.defaultRunOptions()
	ro.ConfigureNetwork = buildah.NetworkEnabled

	if err := b.run(cmd, ro); err != nil {
		return fmt.Errorf("upgrading apk packages: %w", err)
	}

	return nil
}
