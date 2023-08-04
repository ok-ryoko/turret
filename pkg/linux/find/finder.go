// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package find

import (
	"fmt"
	"strings"
)

// Finder is a unique identifier for an implementation of Unix's find utility.
// The zero value represents an unknown implementation.
type Finder int

const (
	BSD = 1 << iota
	BusyBox
	GNU
)

// String returns a string containing the stylized name of the implementation.
func (f Finder) String() string {
	var s string
	switch f {
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

// FinderWrapper wraps Finder to facilitate its parsing from serialized data.
type FinderWrapper struct {
	Finder
}

// UnmarshalText decodes the finder from a string.
func (w *FinderWrapper) UnmarshalText(text []byte) error {
	var err error
	w.Finder, err = parseFinderString(string(text))
	return err
}

func parseFinderString(s string) (Finder, error) {
	var f Finder
	switch strings.ToLower(s) {
	case "bsd":
		f = BSD
	case "busybox":
		f = BusyBox
	case "gnu":
		f = GNU
	default:
		return 0, fmt.Errorf("unsupported finder: %s", s)
	}
	return f, nil
}
