// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package usrgrp

import (
	"fmt"
	"strings"
)

type ShadowCommandFactory struct{}

func (c ShadowCommandFactory) NewCreateUserCmd(name string, options CreateUserOptions) (cmd, capabilities []string) {
	cmd = []string{"useradd", "--create-home"}

	if options.ID > 0 {
		cmd = append(cmd, "--uid", fmt.Sprintf("%d", options.ID))
	}

	if options.UserGroup {
		cmd = append(cmd, "--user-group")
	}

	if options.Comment != nil {
		cmd = append(cmd, "--comment", *options.Comment)
	}

	if options.LoginShell != "" {
		cmd = append(cmd, "--shell", options.LoginShell)
	}

	if len(options.Groups) > 0 {
		cmd = append(cmd, "--groups", strings.Join(options.Groups, ","))
	}

	cmd = append(cmd, name)

	// CAP_DAC_READ_SEARCH and CAP_FSETID are elements of the useradd effective
	// capability set but are not needed for the operation to succeed.
	//
	capabilities = []string{
		"CAP_CHOWN",
		//
		// - Change owner of files copied from /etc/skel to /home/user
		// - Change owner of /var/spool/mail/user

		"CAP_DAC_OVERRIDE",
		//
		// - Open /etc/shadow and /etc/gshadow
		// - Open files copied from /etc/skel to /home/user

		"CAP_FOWNER",
		//
		// - Change owner and mode of temporary files when updating the passwd,
		// shadow, gshadow, group, subuid and subgid files in /etc
		// - Change owner and mode of /home/user and /var/spool/mail/user
		// - Change owner of, set extended attributes on and update timestamps
		// of files copied from /etc/skel to /home/user
	}

	return
}

func (c ShadowCommandFactory) NewAddUserToGroupCmd(user, group string) (cmd, capabilities []string) {
	return []string{}, []string{}
}

func (c ShadowCommandFactory) UserManager() Manager {
	return Shadow
}
