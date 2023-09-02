// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"fmt"

	"github.com/ok-ryoko/turret/pkg/linux/user"
)

type BusyBoxUserFrontend struct {
	UserFrontend
}

func (f *BusyBoxUserFrontend) CreateUser(c *Container, name string, options user.Options) error {
	{
		cmd, capabilities := f.NewCreateUserCmd(name, options)
		ro := c.DefaultRunOptions()
		ro.AddCapabilities = capabilities
		errContext := fmt.Sprintf("creating user using %s", f.Backend())
		if err := c.runWithLogging(cmd, ro, errContext); err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	if options.UserGroup {
		cmd, _ := f.NewAddUserToGroupCmd(name, name)
		ro := c.DefaultRunOptions()
		errContext := fmt.Sprintf("adding user to group using %s", f.Backend())
		if err := c.runWithLogging(cmd, ro, errContext); err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	if len(options.Groups) > 0 {
		for _, g := range options.Groups {
			cmd, _ := f.NewAddUserToGroupCmd(name, g)
			ro := c.DefaultRunOptions()
			errContext := fmt.Sprintf("adding user to group using %s", f.Backend())
			if err := c.runWithLogging(cmd, ro, errContext); err != nil {
				return fmt.Errorf("%w", err)
			}
		}
	}

	return nil
}
