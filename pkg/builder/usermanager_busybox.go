// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"fmt"

	"github.com/ok-ryoko/turret/pkg/linux/usrgrp"
)

type BusyBoxTurretUserManager struct {
	TurretUserManager
}

// CreateUser creates the sole unprivileged user of the working container.
func (um *BusyBoxTurretUserManager) CreateUser(b *TurretBuilder, name string, options usrgrp.CreateUserOptions) error {
	cmd, capabilities := um.NewCreateUserCmd(name, options)
	ro := b.defaultRunOptions()
	ro.AddCapabilities = capabilities
	if err := b.run(cmd, ro); err != nil {
		return fmt.Errorf(
			"creating user using %s: %w",
			um.UserManager().String(),
			err,
		)
	}

	if options.UserGroup {
		cmd, _ = um.NewAddUserToGroupCmd(name, name)
		ro = b.defaultRunOptions()
		if err := b.run(cmd, ro); err != nil {
			return fmt.Errorf(
				"adding user to group using %s: %w",
				um.UserManager().String(),
				err,
			)
		}
	}

	if len(options.Groups) > 0 {
		for _, g := range options.Groups {
			cmd, _ = um.NewAddUserToGroupCmd(name, g)
			ro = b.defaultRunOptions()
			if err := b.run(cmd, ro); err != nil {
				return fmt.Errorf(
					"adding user to group using %s: %w",
					um.UserManager().String(),
					err,
				)
			}
		}
	}

	return nil
}
