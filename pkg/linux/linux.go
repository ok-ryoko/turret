// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package linux

import (
	"fmt"
	"strings"
)

// LinuxDistro is an identifier for an independent Linux-based distribution;
// the zero value represents an unknown distribution
type LinuxDistro int

const (
	Alpine LinuxDistro = 1 << iota
	Arch
	Debian
	Fedora
	OpenSUSE
	Void
)

// DefaultShell returns the known default login shell for the distro
func (d LinuxDistro) DefaultShell() string {
	var s string
	switch d {
	case Alpine:
		s = "/bin/ash"
	case Arch:
		s = "/bin/bash"
	case Debian:
		s = "/bin/bash"
	case Fedora:
		s = "/bin/bash"
	case OpenSUSE:
		s = "/bin/bash"
	case Void:
		s = "/bin/dash"
	default:
		s = ""
	}
	return s
}

// RePackageName returns a regular expression to match valid package names for
// the distro's canonical packaging ecosystem
func (d LinuxDistro) RePackageName() string {
	var p string
	switch d {
	case Alpine:
		p = `^[0-9a-z][+-\._0-9a-z]*[0-9a-z]$`
	case Arch:
		p = `^[0-9a-z][+-\._0-9a-z]*[0-9a-z]$`
	case Debian:
		p = `^[0-9a-z][+-\.0-9a-z]*[0-9a-z]$`
	case Fedora:
		p = `^[0-9A-Za-z][+-\._0-9A-Za-z]*[0-9A-Za-z]$`
	case OpenSUSE:
		p = `^[0-9A-Za-z][+-\._0-9A-Za-z]*[0-9A-Za-z]$`
	case Void:
		p = `^[0-9A-Za-z][+-\._0-9A-Za-z]*[0-9A-Za-z]$`
	default:
		p = ""
	}
	return p
}

// String returns a string containing the stylized name of the distro
func (d LinuxDistro) String() string {
	var s string
	switch d {
	case Alpine:
		s = "Alpine"
	case Arch:
		s = "Arch"
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

// LinuxDistroWrapper wraps LinuxDistro to facilitate its parsing
type LinuxDistroWrapper struct {
	LinuxDistro
}

// UnmarshalText decodes the distro from a string
func (d *LinuxDistroWrapper) UnmarshalText(text []byte) error {
	var err error
	d.LinuxDistro, err = parseDistroString(string(text))
	return err
}

func parseDistroString(s string) (LinuxDistro, error) {
	d, ok := distroStringMap[strings.ToLower(s)]
	if !ok {
		return 0, fmt.Errorf("unsupported distro: %s", s)
	}
	return d, nil
}

var (
	distroStringMap = map[string]LinuxDistro{
		"alpine":   Alpine,
		"arch":     Arch,
		"fedora":   Fedora,
		"debian":   Debian,
		"opensuse": OpenSUSE,
		"void":     Void,
	}
)
