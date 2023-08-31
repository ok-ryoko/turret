// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"fmt"

	"github.com/ok-ryoko/turret/pkg/linux/user"
)

// UserManagerInterface is the interface implemented by a UserManager
// for a particular user and group management utility.
type UserManagerInterface interface {
	// CreateUser creates the sole unprivileged user of the working container.
	CreateUser(c *Container, name string, options user.CreateUserOptions) error
}

// UserManager provides a high-level front end for Buildah for managing
// users and groups in a Linux builder container.
type UserManager struct {
	user.CommandFactory
}

// NewUserManager creates a new UserManager for a particular user and
// group management utility.
func NewUserManager(um user.Manager) (UserManagerInterface, error) {
	cf, err := user.NewCommandFactory(um)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	var tum UserManagerInterface
	switch um {
	case user.BusyBox:
		tum = &BusyBoxUserManager{UserManager{cf}}
	case user.Shadow:
		tum = &ShadowUserManager{UserManager{cf}}
	default:
		return nil, fmt.Errorf("unrecognized user management utility %v", um)
	}
	return tum, nil
}
