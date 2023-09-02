// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package pckg

import (
	"fmt"
	"strings"
)

const (
	APK Backend = 1 << iota
	APT
	DNF
	Pacman
	XBPS
	Zypper
)

// Backend is a unique identifier for a package manager for Linux-based distros.
// The zero value represents an unknown package manager.
type Backend int

// RePackageName returns a regular expression to match valid package names for
// the package manager's ecosystem.
func (b Backend) RePackageName() string {
	var r string
	switch b {
	case APT:
		r = `^[0-9a-z][+\-.0-9a-z]*[0-9a-z]$`
	case APK, Pacman:
		r = `^[0-9a-z][+\-.0-9_a-z]*[0-9a-z]$`
	case DNF, XBPS, Zypper:
		r = `^[0-9A-Za-z][+\-.0-9A-Z_a-z]*[0-9A-Za-z]$`
	default:
		r = ""
	}
	return r
}

// String returns a string containing the stylized name of the package manager.
func (b Backend) String() string {
	var s string
	switch b {
	case APK:
		s = "APK"
	case APT:
		s = "APT"
	case DNF:
		s = "DNF"
	case Pacman:
		s = "Pacman"
	case XBPS:
		s = "XBPS"
	case Zypper:
		s = "Zypper"
	default:
		s = "unknown"
	}
	return s
}

// BackendWrapper wraps Backend to facilitate its parsing from serialized data.
type BackendWrapper struct {
	Backend
}

// UnmarshalText decodes the package manager from a UTF-8-encoded string.
func (w *BackendWrapper) UnmarshalText(text []byte) error {
	var err error
	w.Backend, err = parseBackendString(string(text))
	return err
}

func parseBackendString(s string) (Backend, error) {
	var b Backend
	switch strings.ToLower(s) {
	case "apk":
		b = APK
	case "apt":
		b = APT
	case "dnf":
		b = DNF
	case "pacman":
		b = Pacman
	case "xbps":
		b = XBPS
	case "zypper":
		b = Zypper
	default:
		return 0, fmt.Errorf("unsupported package manager %q", s)
	}
	return b, nil
}
