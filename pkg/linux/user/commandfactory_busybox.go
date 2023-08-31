// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package user

import "fmt"

type BusyBoxCommandFactory struct{}

func (c BusyBoxCommandFactory) NewCreateUserCmd(name string, options CreateUserOptions) (cmd, capabilities []string) {
	cmd = []string{"adduser", "-D"}

	if options.ID > 0 {
		cmd = append(cmd, "-u", fmt.Sprintf("%d", options.ID))
	}

	if options.Comment != nil {
		cmd = append(cmd, "-g", *options.Comment)
	}

	if options.CreateHome {
		cmd = append(cmd, "-h", fmt.Sprintf("/home/%s", name))
	}

	cmd = append(cmd, name)

	// CAP_DAC_OVERRIDE and CAP_FSETID are elements of the adduser effective
	// capability set but are not needed for the operation to succeed
	//
	capabilities = []string{
		"CAP_CHOWN",
		//
		// Change owner of /home/user

		"CAP_FOWNER",
		//
		// Change mode and owner of /home/user as well as temporary files when
		// editing /etc/passwd, /etc/shadow and /etc/group
	}

	return cmd, capabilities
}

func (c BusyBoxCommandFactory) NewAddUserToGroupCmd(name string, group string) (cmd, capabilities []string) {
	cmd = []string{"addgroup", name, group}
	return cmd, []string{}
}

func (c BusyBoxCommandFactory) UserManager() Manager {
	return BusyBox
}
