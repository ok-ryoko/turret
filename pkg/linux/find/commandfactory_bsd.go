// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package find

type BSDCommandFactory struct{}

func (c BSDCommandFactory) NewFindSpecialCmd() (cmd, capabilities []string) {
	cmd = []string{
		"find", "-x",
		"/",
		"-perm", "+u=s,g=s",
	}
	capabilities = []string{"CAP_DAC_READ_SEARCH"}
	return cmd, capabilities
}
