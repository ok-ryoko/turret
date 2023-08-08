// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package usrgrp

import "fmt"

// CommandFactory provides a simple layer of abstraction over common user and
// group management operations.
type CommandFactory interface {
	// NewAddUserToGroupCmd returns (1) a command that adds a user to a group
	// and (2) the Linux capabilities needed by that command.
	NewAddUserToGroupCmd(user, group string) (cmd, capabilities []string)

	// NewCreateUserCmd returns (1) a command that creates a new user and (2)
	// the Linux capabilities needed by that command.
	NewCreateUserCmd(name string, options CreateUserOptions) (cmd, capabilities []string)

	// UserManager returns a constant representing the user/group management
	// utility for which this factory makes commands.
	UserManager() Manager
}

// NewCommandFactory creates a new CommandFactory that manufactures user and
// group management commands for execution in a shell.
func NewCommandFactory(m Manager) (CommandFactory, error) {
	var result CommandFactory
	switch m {
	case BusyBox:
		result = &BusyBoxCommandFactory{}
	case Shadow:
		result = &ShadowCommandFactory{}
	default:
		return nil, fmt.Errorf("unrecognized user/group management utility")
	}
	return result, nil
}

type CreateUserOptions struct {
	ID         uint
	Comment    *string
	UserGroup  bool
	Groups     []string
	CreateHome bool
	Shell      string
}
