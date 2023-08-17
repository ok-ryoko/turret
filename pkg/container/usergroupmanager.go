// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"fmt"

	"github.com/ok-ryoko/turret/pkg/linux/usrgrp"
)

// UserGroupManagerInterface is the interface implemented by a UserGroupManager
// for a particular user and group management utility.
type UserGroupManagerInterface interface {
	// CreateUser creates the sole unprivileged user of the working container.
	CreateUser(c *Container, name string, options usrgrp.CreateUserOptions) error
}

// UserGroupManager provides a high-level front end for Buildah for managing
// users and groups in a Linux builder container.
type UserGroupManager struct {
	usrgrp.CommandFactory
}

// NewUserGroupManager creates a new UserGroupManager for a particular user and
// group management utility.
func NewUserGroupManager(um usrgrp.Manager) (UserGroupManagerInterface, error) {
	cf, err := usrgrp.NewCommandFactory(um)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	var tum UserGroupManagerInterface
	switch um {
	case usrgrp.BusyBox:
		tum = &BusyBoxUserGroupManager{UserGroupManager{cf}}
	case usrgrp.Shadow:
		tum = &ShadowUserGroupManager{UserGroupManager{cf}}
	default:
		return nil, fmt.Errorf("unrecognized user management utility %v", um)
	}
	return tum, nil
}
