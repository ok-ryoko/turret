// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"fmt"

	"github.com/ok-ryoko/turret/pkg/linux"
)

type AlpineTurretBuilder struct {
	TurretBuilder
}

func (b *AlpineTurretBuilder) CreateUser(name string, distro linux.Distro, options CreateUserOptions) error {
	if name == "" {
		return fmt.Errorf("blank user name")
	}

	adduserCmd := []string{"adduser", "-D"}

	if !options.UserGroup {
		adduserCmd = append(adduserCmd, "-G", "users")
	}

	if options.LoginShell != distro.DefaultShell() {
		shell, err := b.resolveExecutable(options.LoginShell, distro)
		if err != nil {
			return fmt.Errorf("resolving login shell: %w", err)
		}
		options.LoginShell = shell
	}
	adduserCmd = append(adduserCmd, "-s", options.LoginShell)

	if options.ID != 0 {
		if options.ID < 1000 || options.ID > 60000 {
			return fmt.Errorf("UID %d outside allowed range [1000-60000]", options.ID)
		}
		adduserCmd = append(
			adduserCmd,
			"-u",
			fmt.Sprintf("%d", options.ID),
		)
	}

	if options.Comment != "" {
		adduserCmd = append(adduserCmd, "-g", options.Comment)
	}

	adduserCmd = append(adduserCmd, name)

	// CAP_DAC_OVERRIDE and CAP_FSETID are elements of the useradd effective
	// capability set but are not needed for the operation to succeed
	//
	auo := b.defaultRunOptions()
	auo.AddCapabilities = []string{
		"CAP_CHOWN",
		//
		// Change owner of /home/user

		"CAP_FOWNER",
		//
		// Change mode and owner of /home/user as well as temporary files when
		// editing /etc/passwd, /etc/shadow and /etc/group
	}

	if err := b.run(adduserCmd, auo); err != nil {
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
