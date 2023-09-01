// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"fmt"
	"strings"
)

const (
	BusyBox Manager = 1 << iota
	Shadow
)

// Manager is a unique identifier for a user/group management utility for
// Linux-based distros. The zero value represents an unknown program.
type Manager int

// String returns a string containing the stylized name of the user/group
// management utility.
func (m Manager) String() string {
	var s string
	switch m {
	case BusyBox:
		s = "BusyBox"
	case Shadow:
		s = "shadow-utils"
	default:
		s = "unknown"
	}
	return s
}

// ManagerWrapper wraps Manager to facilitate its parsing from serialized data.
type ManagerWrapper struct {
	Manager
}

// UnmarshalText decodes the user/group management utility from a string.
func (w *ManagerWrapper) UnmarshalText(text []byte) error {
	var err error
	w.Manager, err = parseManagerString(string(text))
	return err
}

func parseManagerString(s string) (Manager, error) {
	var m Manager
	switch strings.ToLower(s) {
	case "busybox":
		m = BusyBox
	case "shadow", "shadow-utils":
		m = Shadow
	default:
		return 0, fmt.Errorf("unsupported user management utility %q", s)
	}
	return m, nil
}
