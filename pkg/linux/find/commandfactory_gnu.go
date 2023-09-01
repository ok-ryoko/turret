// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package find

type GNUCommandFactory struct{}

func (f GNUCommandFactory) NewFindSpecialCmd() (cmd, capabilities []string) {
	cmd = []string{
		"find", "/",
		"-xdev",
		"-perm", "/u=s,g=s",
	}
	capabilities = []string{"CAP_DAC_READ_SEARCH"}
	return cmd, capabilities
}
