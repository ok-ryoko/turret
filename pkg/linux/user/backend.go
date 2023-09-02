// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"fmt"
	"strings"
)

const (
	BusyBox Backend = 1 << iota
	Shadow
)

// Backend is a unique identifier for a user and group management utility for
// Linux-based distros. The zero value represents an unknown utility.
type Backend int

// String returns a string containing the stylized name of the user and group
// management utility.
func (b Backend) String() string {
	var s string
	switch b {
	case BusyBox:
		s = "BusyBox"
	case Shadow:
		s = "shadow-utils"
	default:
		s = "unknown"
	}
	return s
}

// BackendWrapper wraps Backend to facilitate its parsing from serialized data.
type BackendWrapper struct {
	Backend
}

// UnmarshalText decodes the user and group management utility from a
// UTF-8-encoded string.
func (w *BackendWrapper) UnmarshalText(text []byte) error {
	var err error
	w.Backend, err = parseBackendString(string(text))
	return err
}

func parseBackendString(s string) (Backend, error) {
	var b Backend
	switch strings.ToLower(s) {
	case "busybox":
		b = BusyBox
	case "shadow", "shadow-utils":
		b = Shadow
	default:
		return 0, fmt.Errorf("unsupported user management utility %q", s)
	}
	return b, nil
}
