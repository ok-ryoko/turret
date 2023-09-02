// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package find

import "fmt"

// CommandFactory provides a layer of abstraction over search operations.
type CommandFactory interface {
	// NewFindSpecialCmd returns (1) a command that finds all files with a
	// SUID and/or SGID bit in real (non-device) file systems, and (2) the
	// Linux capabilities needed by that command.
	NewFindSpecialCmd() (cmd, capabilities []string)
}

// NewCommandFactory creates an object that manufactures find commands for
// execution in a shell.
func NewCommandFactory(b Backend) (CommandFactory, error) {
	var factory CommandFactory
	switch b {
	case BSD:
		factory = &BSDCommandFactory{}
	case BusyBox:
		factory = &BusyBoxCommandFactory{}
	case GNU:
		factory = &GNUCommandFactory{}
	default:
		return nil, fmt.Errorf("unrecognized find implementation %v", b)
	}
	return factory, nil
}
