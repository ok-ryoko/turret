// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"fmt"

	"github.com/ok-ryoko/turret/pkg/linux/user"
)

// UserFrontendInterface is the interface implemented by a UserFrontend for a
// particular user and group management backend.
type UserFrontendInterface interface {
	// CreateUser creates the sole unprivileged user of the working container.
	CreateUser(c *Container, name string, options user.Options) error
}

// UserFrontend provides a high-level frontend for Buildah for managing users
// and groups in a Linux builder container.
type UserFrontend struct {
	user.CommandFactory
}

// NewUserFrontend creates a frontend for a particular user and group management
// backend.
func NewUserFrontend(backend user.Backend) (UserFrontendInterface, error) {
	factory, err := user.NewCommandFactory(backend)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	var frontend UserFrontendInterface
	switch backend {
	case user.BusyBox:
		frontend = &BusyBoxUserFrontend{UserFrontend{factory}}
	case user.Shadow:
		frontend = &ShadowUserFrontend{UserFrontend{factory}}
	default:
		return nil, fmt.Errorf("unrecognized user and group management backend %v", backend)
	}
	return frontend, nil
}
