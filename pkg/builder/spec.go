// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	rePOSIXPortableName    string = `^[0-9A-Za-z]$|^[0-9A-Za-z][-\._0-9A-Za-z]*[0-9A-Za-z]$`
	reInvalidName          string = `^[0-9]*$|^\.{1,2}$`
	reReverseUnlimitedFQDN string = `^\.?([0-9A-Za-z]|[0-9A-Za-z][-0-9A-Za-z]*[0-9A-Za-z]\.)*[0-9A-Za-z]$`
)

// Spec holds the options for the build and defines the structure of spec files
type Spec struct {
	// GNU/Linux distro for this image; must match the distro in the base image
	Distro GNULinuxDistroWrapper

	// Fully qualified name for the image we're building
	Repository string

	// Tag for the image we're building
	Tag string

	// Reference for the base image
	From BaseImage

	// Whether to upgrade the packages already installed in the base image
	Upgrade bool

	// List of packages to install using the distro's canonical package manager
	Packages []string

	// Whether to clean package caches
	Clean bool

	// Sole unprivileged user in the working container
	User User

	// Map of destination paths in the working container's file system to
	// sources in the end user's home directory
	Copy map[string][]string

	// Map of environment variables to set or update in the working container
	Env map[string]string

	// Map of annotations to apply to the working container
	Annotations map[string]string

	// Whether to preserve the image history and timestamps of the files in the
	// working container's file system
	KeepHistory bool `toml:"keep-history"`

	// Security options for the working container
	Security Security
}

// Fill populates the spec using runtime information
func (s *Spec) Fill() {
	if s.User.LoginShell == "" {
		s.User.LoginShell = s.Distro.DefaultShell()
	}
}

// Validate asserts that the spec is suitable for ingestion by a builder
func (s *Spec) Validate() error {
	if (s.Distro == GNULinuxDistroWrapper{0}) {
		return fmt.Errorf("missing distro")
	}

	if s.Repository == "" {
		return fmt.Errorf("missing image name")
	}

	if s.From.Repository == "" || s.From.Tag == "" {
		return fmt.Errorf("incomplete base image specification (missing repository or tag)")
	}

	if len(s.Packages) > 0 {
		re, err := regexp.Compile(s.Distro.RePackageName())
		if err != nil {
			return fmt.Errorf("compiling regular expression: %w", err)
		}
		for _, p := range s.Packages {
			if !re.MatchString(p) {
				return fmt.Errorf("invalid package name '%s'", p)
			}
		}
	}

	if s.User.Create {
		re1, err := regexp.Compile(rePOSIXPortableName)
		if err != nil {
			return fmt.Errorf("compiling regular expression: %w", err)
		}

		re2, err := regexp.Compile(reInvalidName)
		if err != nil {
			return fmt.Errorf("compiling regular expression: %w", err)
		}

		if s.User.Name != "" {
			if s.User.Name == "root" {
				return fmt.Errorf("won't create unprivileged user named 'root'")
			}

			if len(s.User.Name) > 32 {
				return fmt.Errorf("user name '%s' too long (limit: 32 chars)", s.User.Name)
			}

			if !re1.MatchString(s.User.Name) || re2.MatchString(s.User.Name) {
				return fmt.Errorf("invalid user name '%s'", s.User.Name)
			}
		}

		if s.User.UID < 1000 || s.User.UID > 60000 {
			return fmt.Errorf("UID %d outside allowed range [1000-60000]", s.User.UID)
		}

		if len(s.User.Groups) > 0 {
			for _, g := range s.User.Groups {
				if !re1.MatchString(g) || re2.MatchString(g) {
					return fmt.Errorf("invalid group name '%s'", g)
				}
			}
		}
	}

	if len(s.Annotations) > 0 {
		r, err := regexp.Compile(reReverseUnlimitedFQDN)
		if err != nil {
			return fmt.Errorf("compiling regular expression: %w", err)
		}
		for k := range s.Annotations {
			if !r.MatchString(k) {
				return fmt.Errorf("invalid annotation key '%s'", k)
			}
		}
	}

	return nil
}

// NewSpec generates a default but invalid spec
func NewSpec() Spec {
	return Spec{
		Distro:      GNULinuxDistroWrapper{0},
		Repository:  "",
		Tag:         "",
		From:        BaseImage{},
		Upgrade:     false,
		Packages:    []string{},
		Clean:       false,
		User:        DefaultUser(),
		Copy:        map[string][]string{},
		Env:         map[string]string{},
		Annotations: map[string]string{},
		KeepHistory: false,
		Security:    Security{},
	}
}

// GNULinuxDistro is an identifier for an independent GNU/Linux distribution;
// the zero value represents an unknown distribution
type GNULinuxDistro int

const (
	Alpine GNULinuxDistro = 1 << iota
	Arch
	Debian
	Fedora
	OpenSUSE
	Void
)

// DefaultShell returns the known default login shell for the distro
func (d GNULinuxDistro) DefaultShell() string {
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
func (d GNULinuxDistro) RePackageName() string {
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
func (d GNULinuxDistro) String() string {
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

// GNULinuxDistroWrapper wraps GNULinuxDistro to facilitate its parsing
type GNULinuxDistroWrapper struct {
	GNULinuxDistro
}

// UnmarshalText decodes the distro from a string
func (d *GNULinuxDistroWrapper) UnmarshalText(text []byte) error {
	var err error
	d.GNULinuxDistro, err = parseDistroString(string(text))
	return err
}

func parseDistroString(s string) (GNULinuxDistro, error) {
	d, ok := distroStringMap[strings.ToLower(s)]
	if !ok {
		return 0, fmt.Errorf("unsupported distro: %s", s)
	}
	return d, nil
}

var (
	distroStringMap = map[string]GNULinuxDistro{
		"alpine":   Alpine,
		"arch":     Arch,
		"fedora":   Fedora,
		"debian":   Debian,
		"opensuse": OpenSUSE,
		"void":     Void,
	}
)

// BaseImage holds the components of the base image reference
type BaseImage struct {
	Repository string
	Tag        string
}

// Reference returns the reference of the image in the repository:tag format
func (i BaseImage) Reference() string {
	return fmt.Sprintf("%s:%s", i.Repository, i.Tag)
}

// User holds information about a Linux user
type User struct {
	// Whether to create this user
	Create bool

	// Human-readable identifier
	Name string

	// Numeric identifier from 1000 to 60000, inclusive
	UID uint `toml:"uid"`

	// Whether to create a user group
	UserGroup bool `toml:"user-group"`

	// Groups to which to add the user
	Groups []string

	// GECOS field text
	Comment string

	// Login shell; must be a PATH-resolvable executable
	LoginShell string `toml:"login-shell"`
}

// DefaultUser returns a minimally configured Linux user
func DefaultUser() User {
	return User{
		Create:     false,
		Name:       "user",
		UID:        1000,
		UserGroup:  false,
		Groups:     []string{},
		Comment:    "",
		LoginShell: "",
	}
}

// Security holds security-related options for the working container
type Security struct {
	// Options for handling files with a SUID/SGID bit
	SpecialFiles SpecialFiles `toml:"special-files"`
}

// SpecialFiles holds options for handling SUID/SGID bits
type SpecialFiles struct {
	// Whether to remove all SUID/SGID bits automatically
	RemoveS bool `toml:"remove-s"`

	// Whether to preserve the SUID/SGID bit on one or more files
	Excludes []string
}
