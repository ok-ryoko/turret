// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package find

import (
	"fmt"
	"strings"
)

const (
	BSD Backend = 1 << iota
	BusyBox
	GNU
)

// Backend is a unique identifier for an implementation of Unix's find utility.
// The zero value represents an unknown implementation.
type Backend uint

// String returns a string containing the stylized name of the implementation.
func (b Backend) String() string {
	var s string
	switch b {
	case BSD:
		s = "BSD"
	case BusyBox:
		s = "BusyBox"
	case GNU:
		s = "GNU"
	default:
		s = "unknown"
	}
	return s
}

// BackendWrapper wraps Backend to facilitate its parsing from serialized data.
type BackendWrapper struct {
	Backend
}

// UnmarshalText decodes the finder from a UTF-8-encoded string.
func (w *BackendWrapper) UnmarshalText(text []byte) error {
	var err error
	w.Backend, err = parseBackendString(string(text))
	return err
}

func parseBackendString(s string) (Backend, error) {
	var b Backend
	switch strings.ToLower(s) {
	case "bsd":
		b = BSD
	case "busybox":
		b = BusyBox
	case "gnu":
		b = GNU
	default:
		return 0, fmt.Errorf("unsupported find implementation %q", s)
	}
	return b, nil
}
