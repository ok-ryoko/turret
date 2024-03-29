// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package find

type BusyBoxCommandFactory struct{}

func (f BusyBoxCommandFactory) NewFindSpecialCmd() (cmd, capabilities []string) {
	cmd = []string{
		"find", "/",
		"-xdev",
		"(", "-perm", "+2000", "-o", "-perm", "+4000", ")",
	}
	return cmd, []string{}
}
