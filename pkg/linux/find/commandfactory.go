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
func NewCommandFactory(f Finder) (CommandFactory, error) {
	var result CommandFactory
	switch f {
	case BSD:
		result = &BSDCommandFactory{}
	case BusyBox:
		result = &BusyBoxCommandFactory{}
	case GNU:
		result = &GNUCommandFactory{}
	default:
		return nil, fmt.Errorf("unrecognized find implementation %v", f)
	}
	return result, nil
}
