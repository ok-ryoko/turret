// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package linux

import (
	"fmt"
	"strings"

	"github.com/ok-ryoko/turret/pkg/linux/find"
	"github.com/ok-ryoko/turret/pkg/linux/pckg"
	"github.com/ok-ryoko/turret/pkg/linux/user"
)

const (
	Alpine Distro = 1 << iota
	Arch
	Chimera
	Debian
	Fedora
	OpenSUSE
	Void
)

// Distro is a unique identifier for an independent Linux-based distribution.
// The zero value represents an unknown distro.
type Distro int

// DefaultPackageManager returns the canonical package manager for the distro.
func (d Distro) DefaultPackageManager() pckg.Manager {
	var m pckg.Manager
	switch d {
	case Alpine, Chimera:
		m = pckg.APK
	case Arch:
		m = pckg.Pacman
	case Debian:
		m = pckg.APT
	case Fedora:
		m = pckg.DNF
	case OpenSUSE:
		m = pckg.Zypper
	case Void:
		m = pckg.XBPS
	default:
		m = 0
	}
	return m
}

// DefaultUserManager returns the canonical user and group management utility
// for the distro.
func (d Distro) DefaultUserManager() user.Manager {
	var m user.Manager
	switch d {
	case Alpine:
		m = user.BusyBox
	case Arch, Chimera, Debian, Fedora, OpenSUSE, Void:
		m = user.Shadow
	default:
		m = 0
	}
	return m
}

// DefaultFinder returns the canonical implementation of the find utility for
// the distro.
func (d Distro) DefaultFinder() find.Finder {
	var f find.Finder
	switch d {
	case Alpine:
		f = find.BusyBox
	case Chimera:
		f = find.BSD
	case Arch, Debian, Fedora, OpenSUSE, Void:
		f = find.GNU
	default:
		f = 0
	}
	return f
}

// String returns a string containing the stylized name of the distro.
func (d Distro) String() string {
	var s string
	switch d {
	case Alpine:
		s = "Alpine"
	case Arch:
		s = "Arch"
	case Chimera:
		s = "Chimera"
	case Debian:
		s = "Debian"
	case Fedora:
		s = "Fedora"
	case OpenSUSE:
		s = "openSUSE"
	case Void:
		s = "Void"
	default:
		s = "unknown"
	}
	return s
}

// DistroWrapper wraps Distro to facilitate its parsing from serialized data.
type DistroWrapper struct {
	Distro
}

// UnmarshalText decodes the distro from a UTF-8-encoded string.
func (w *DistroWrapper) UnmarshalText(text []byte) error {
	var err error
	w.Distro, err = parseDistroString(string(text))
	return err
}

func parseDistroString(s string) (Distro, error) {
	var d Distro
	switch strings.ToLower(s) {
	case "alpine":
		d = Alpine
	case "arch":
		d = Arch
	case "chimera":
		d = Chimera
	case "debian":
		d = Debian
	case "fedora":
		d = Fedora
	case "opensuse":
		d = OpenSUSE
	case "void":
		d = Void
	default:
		return 0, fmt.Errorf("unsupported distro %q", s)
	}
	return d, nil
}
