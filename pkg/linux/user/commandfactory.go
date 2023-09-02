// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package user

import "fmt"

// CommandFactory provides a layer of abstraction over user and group management
// operations.
type CommandFactory interface {
	// NewAddUserToGroupCmd returns (1) a command that adds a user to a group
	// and (2) the Linux capabilities needed by that command.
	NewAddUserToGroupCmd(user, group string) (cmd, capabilities []string)

	// NewCreateUserCmd returns (1) a command that creates a new user and (2)
	// the Linux capabilities needed by that command.
	NewCreateUserCmd(name string, options Options) (cmd, capabilities []string)

	// UserManager returns a constant representing the user and group management
	// utility for which this factory makes commands.
	UserManager() Manager
}

// Options holds options for a Linux user.
type Options struct {
	// Linux user ID
	ID uint32

	// Create a user group
	UserGroup bool

	// Groups to which to add the user
	Groups []string

	// GECOS field text
	Comment *string

	// Create a home directory for the user in /home
	CreateHome bool
}

// NewCommandFactory creates an object that manufactures user and group
// management commands for execution in a shell.
func NewCommandFactory(m Manager) (CommandFactory, error) {
	var result CommandFactory
	switch m {
	case BusyBox:
		result = &BusyBoxCommandFactory{}
	case Shadow:
		result = &ShadowCommandFactory{}
	default:
		return nil, fmt.Errorf("unrecognized user management utility %v", m)
	}
	return result, nil
}
