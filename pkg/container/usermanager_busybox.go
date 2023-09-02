// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"fmt"

	"github.com/ok-ryoko/turret/pkg/linux/user"
)

type BusyBoxUserManager struct {
	UserManager
}

func (m *BusyBoxUserManager) CreateUser(c *Container, name string, options user.CreateUserOptions) error {
	{
		cmd, capabilities := m.NewCreateUserCmd(name, options)
		ro := c.DefaultRunOptions()
		ro.AddCapabilities = capabilities
		errContext := fmt.Sprintf("creating user using %s", m.UserManager.UserManager())
		if err := c.runWithLogging(cmd, ro, errContext); err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	if options.UserGroup {
		cmd, _ := m.NewAddUserToGroupCmd(name, name)
		ro := c.DefaultRunOptions()
		errContext := fmt.Sprintf("adding user to group using %s", m.UserManager.UserManager())
		if err := c.runWithLogging(cmd, ro, errContext); err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	if len(options.Groups) > 0 {
		for _, g := range options.Groups {
			cmd, _ := m.NewAddUserToGroupCmd(name, g)
			ro := c.DefaultRunOptions()
			errContext := fmt.Sprintf("adding user to group using %s", m.UserManager.UserManager())
			if err := c.runWithLogging(cmd, ro, errContext); err != nil {
				return fmt.Errorf("%w", err)
			}
		}
	}

	return nil
}
