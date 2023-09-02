// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"fmt"

	"github.com/ok-ryoko/turret/pkg/linux/user"
)

type ShadowUserManager struct {
	UserManager
}

func (m *ShadowUserManager) CreateUser(c *Container, name string, options user.Options) error {
	cmd, capabilities := m.NewCreateUserCmd(name, options)
	ro := c.DefaultRunOptions()
	ro.AddCapabilities = capabilities

	// If the sss_cache command is available, then useradd will fork into
	// sss_cache to invalidate the System Security Services Daemon cache,
	// an operation that requires additional capabilities.
	//
	_, err := c.ResolveExecutable("sss_cache")
	if err != nil {
		c.Logger.Debugln("sss_cache not found; skipping cache invalidation")
	} else {
		ro.AddCapabilities = append(
			ro.AddCapabilities,
			"CAP_SETGID",
			//
			// Set the effective GID to 0 (root)

			"CAP_SETUID",
			//
			// Set the effective UID to 0 (root)
		)
	}

	errContext := fmt.Sprintf("creating user using %s", m.UserManager.UserManager())
	if err := c.runWithLogging(cmd, ro, errContext); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}
