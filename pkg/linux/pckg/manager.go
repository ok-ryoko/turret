// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package pckg

import (
	"fmt"
	"strings"
)

const (
	APK Manager = 1 << iota
	APT
	DNF
	Pacman
	XBPS
	Zypper
)

// Manager is a unique identifier for a package manager for Linux-based
// distros. The zero value represents an unknown package manager.
type Manager int

// RePackageName returns a regular expression to match valid package names for
// the package manager's ecosystem.
func (m Manager) RePackageName() string {
	var r string
	switch m {
	case APT:
		r = `^[0-9a-z][+-\.0-9a-z]*[0-9a-z]$`
	case APK, Pacman:
		r = `^[0-9a-z][+-\._0-9a-z]*[0-9a-z]$`
	case DNF, XBPS, Zypper:
		r = `^[0-9A-Za-z][+-\._0-9A-Za-z]*[0-9A-Za-z]$`
	default:
		r = ""
	}
	return r
}

// String returns a string containing the stylized name of the package manager.
func (m Manager) String() string {
	var s string
	switch m {
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

// ManagerWrapper wraps Manager to facilitate its parsing from serialized data.
type ManagerWrapper struct {
	Manager
}

// UnmarshalText decodes the package manager from a string.
func (mw *ManagerWrapper) UnmarshalText(text []byte) error {
	var err error
	mw.Manager, err = parseManagerString(string(text))
	return err
}

func parseManagerString(s string) (Manager, error) {
	var m Manager
	switch strings.ToLower(s) {
	case "apk":
		m = APK
	case "apt":
		m = APT
	case "dnf":
		m = DNF
	case "pacman":
		m = Pacman
	case "xbps":
		m = XBPS
	case "zypper":
		m = Zypper
	default:
		return 0, fmt.Errorf("unsupported package manager %q", s)
	}
	return m, nil
}
