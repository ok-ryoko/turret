// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package linux

import (
	"fmt"
	"strings"

	"github.com/ok-ryoko/turret/pkg/linux/packagemanager"
)

// Distro is a unique identifier for an independent Linux-based distribution.
// The zero value represents an unknown distro.
type Distro int

const (
	Alpine Distro = 1 << iota
	Arch
	Chimera
	Debian
	Fedora
	OpenSUSE
	Void
)

// DefaultShell returns the known default login shell for the distro.
func (d Distro) DefaultShell() string {
	var s string
	switch d {
	case Alpine:
		s = "/bin/ash"
	case Arch, Debian, Fedora, OpenSUSE:
		s = "/bin/bash"
	case Chimera:
		s = "/bin/sh"
	case Void:
		s = "/bin/dash"
	default:
		s = ""
	}
	return s
}

// DefaultPackageManager returns the canonical package manager for the distro.
func (d Distro) DefaultPackageManager() packagemanager.PackageManager {
	var p packagemanager.PackageManager
	switch d {
	case Alpine, Chimera:
		p = packagemanager.APK
	case Arch:
		p = packagemanager.Pacman
	case Debian:
		p = packagemanager.APT
	case Fedora:
		p = packagemanager.DNF
	case OpenSUSE:
		p = packagemanager.Zypper
	case Void:
		p = packagemanager.XBPS
	default:
		p = 0
	}
	return p
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

// UnmarshalText decodes the distro from a string.
func (d *DistroWrapper) UnmarshalText(text []byte) error {
	var err error
	d.Distro, err = parseDistroString(string(text))
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
		return 0, fmt.Errorf("unsupported distro: %s", s)
	}
	return d, nil
}
