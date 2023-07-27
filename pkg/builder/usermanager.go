// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"fmt"

	"github.com/ok-ryoko/turret/pkg/linux/usrgrp"
)

// TurretUserManagerInterface is the interface implemented by a
// TurretUserManager for a particular user and group management utility.
type TurretUserManagerInterface interface {
	// CreateUser creates the sole unprivileged user of the working container.
	CreateUser(c *TurretContainer, name string, options usrgrp.CreateUserOptions) error
}

// TurretUserManager provides a high-level front end for Buildah for managing
// users and groups in a Linux builder container.
type TurretUserManager struct {
	usrgrp.CommandFactory
}

// NewUserManager creates a new TurretUserManager for a particular user and
// group management utility.
func NewUserManager(um usrgrp.Manager) (TurretUserManagerInterface, error) {
	cf, err := usrgrp.NewCommandFactory(um)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	var tum TurretUserManagerInterface
	switch um {
	case usrgrp.BusyBox:
		tum = &BusyBoxTurretUserManager{TurretUserManager{cf}}
	case usrgrp.Shadow:
		tum = &ShadowTurretUserManager{TurretUserManager{cf}}
	default:
		return nil, fmt.Errorf("unrecognized package manager")
	}
	return tum, nil
}
