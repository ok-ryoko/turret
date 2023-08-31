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

// CreateUser creates the sole unprivileged user of the working container.
func (um *BusyBoxUserManager) CreateUser(c *Container, name string, options user.CreateUserOptions) error {
	{
		cmd, capabilities := um.NewCreateUserCmd(name, options)
		ro := c.DefaultRunOptions()
		ro.AddCapabilities = capabilities
		errMsg := fmt.Sprintf("creating user using %s", um.UserManager.UserManager())
		if err := c.runWithLogging(cmd, ro, errMsg); err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	if options.UserGroup {
		cmd, _ := um.NewAddUserToGroupCmd(name, name)
		ro := c.DefaultRunOptions()
		errMsg := fmt.Sprintf("adding user to group using %s", um.UserManager.UserManager())
		if err := c.runWithLogging(cmd, ro, errMsg); err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	if len(options.Groups) > 0 {
		for _, g := range options.Groups {
			cmd, _ := um.NewAddUserToGroupCmd(name, g)
			ro := c.DefaultRunOptions()
			errMsg := fmt.Sprintf("adding user to group using %s", um.UserManager.UserManager())
			if err := c.runWithLogging(cmd, ro, errMsg); err != nil {
				return fmt.Errorf("%w", err)
			}
		}
	}

	return nil
}
