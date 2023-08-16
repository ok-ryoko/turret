// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ok-ryoko/turret/pkg/linux"
	"github.com/ok-ryoko/turret/pkg/linux/find"
	"github.com/ok-ryoko/turret/pkg/linux/pckg"
	"github.com/ok-ryoko/turret/pkg/linux/usrgrp"
)

var (
	reDigits                    = regexp.MustCompile(`^[0-9]+$`)
	reNotPOSIXPortableCharacter = regexp.MustCompile(`[^-.0-9A-Z_a-z]`)
	reReverseUnlimitedFQDN      = regexp.MustCompile(`^\.?([0-9A-Za-z]|[0-9A-Za-z][-0-9A-Za-z]*[0-9A-Za-z]\.)*[0-9A-Za-z]$`)
	reSpecialPrefixOrSuffix     = regexp.MustCompile(`^[-._]|[-._]$`)
	reURLScheme                 = regexp.MustCompile(`^[^:/?#]+:`) // IETF RFC 3986 Appendix B
)

// Spec holds the options for the build and defines the structure of spec files.
type Spec struct {
	// Linux-based distro for this image
	Distro linux.DistroWrapper

	// Fully qualified name for the image we're building
	Repository string

	// Tag for the image we're building
	Tag string

	// Reference for the base image
	From BaseImage

	// Instructions for the distro's canonical package manager
	Packages Packages

	// Sole unprivileged user in the working container
	User *User

	// Instructions and options for copying one or more files from the host's
	// file system to the working container's file system
	Copy []Copy

	// Whether to preserve the image history and timestamps of the files in the
	// working container's file system
	KeepHistory bool `toml:"keep-history"`

	// Security options for the working container
	Security Security

	// Configuration for the working container
	Config Configuration

	// Choices of implementations of operations in the working container
	Backends Backends
}

// Fill populates empty optional fields in a spec using information encoded
// by required fields in the spec.
func Fill(s Spec) Spec {
	if s.Backends.Package.Manager == 0 {
		s.Backends.Package.Manager = s.Distro.DefaultPackageManager()
	}

	if s.Backends.User.Manager == 0 {
		s.Backends.User.Manager = s.Distro.DefaultUserManager()
	}

	if s.Backends.Finder.Finder == 0 {
		s.Backends.Finder.Finder = s.Distro.DefaultFinder()
	}

	if s.Config.Annotations == nil {
		s.Config.Annotations = map[string]string{}
	}

	if s.Config.Environment == nil {
		s.Config.Environment = map[string]string{}
	}

	if s.Config.Labels == nil {
		s.Config.Labels = map[string]string{}
	}

	for i, p := range s.Config.Ports {
		if p.Protocol.Protocol == 0 {
			s.Config.Ports[i].Protocol.Protocol = TCP
		}
	}

	return s
}

// Validate asserts that a spec is suitable for ingestion by a builder.
func Validate(s Spec) error {
	if s.Distro.Distro == 0 {
		return fmt.Errorf("missing distro")
	}

	if s.Backends.Package.Manager == 0 {
		return fmt.Errorf("missing package manager")
	}

	if s.Backends.User.Manager == 0 {
		return fmt.Errorf("missing user/group management utility")
	}

	if s.Backends.Finder.Finder == 0 {
		return fmt.Errorf("missing find implementation")
	}

	if s.Repository == "" {
		return fmt.Errorf("missing image repository (name)")
	}

	if s.From.Repository == "" || s.From.Tag == "" {
		return fmt.Errorf("missing base image repository (name) or tag")
	}

	if len(s.Packages.Install) > 0 {
		re := regexp.MustCompile(s.Backends.Package.RePackageName())
		for _, p := range s.Packages.Install {
			if !re.MatchString(p) {
				return fmt.Errorf("invalid package name '%s'", p)
			}
		}
	}

	if s.User != nil {
		err := validateName(s.User.Name)
		if err != nil {
			return fmt.Errorf("invalid user name '%s': %w", s.User.Name, err)
		}

		if s.User.ID != 0 && (s.User.ID < 1000 || s.User.ID > 60000) {
			return fmt.Errorf("UID %d outside allowed range [1000-60000]", s.User.ID)
		}

		for _, g := range s.User.Groups {
			err := validateName(g)
			if err != nil {
				return fmt.Errorf("invalid group name '%s': %w", g, err)
			}
		}
	}

	for _, c := range s.Copy {
		if c.Base == "" {
			return fmt.Errorf("missing base")
		}
		if !filepath.IsAbs(c.Base) {
			return fmt.Errorf("base is not an absolute path")
		}

		if c.Destination == "" {
			return fmt.Errorf("missing destination")
		}
		if !filepath.IsAbs(c.Destination) {
			return fmt.Errorf("destination is not an absolute path")
		}

		if len(c.Sources) == 0 {
			return fmt.Errorf("missing sources for destination %q", c.Destination)
		}
		for i, src := range c.Sources {
			if src == "" {
				return fmt.Errorf("empty source at index %d for destination %q", i, c.Destination)
			}
			if reURLScheme.MatchString(src) {
				return fmt.Errorf("only schemeless paths are supported (%q)", src)
			}
		}
	}

	if a := s.Config.Annotations; len(a) > 0 {
		for k := range a {
			if !reReverseUnlimitedFQDN.MatchString(k) {
				return fmt.Errorf("invalid annotation key: %q", k)
			}
		}
	}

	if l := s.Config.Labels; len(l) > 0 {
		for k := range l {
			if !reReverseUnlimitedFQDN.MatchString(k) {
				return fmt.Errorf("invalid label key: %q", k)
			}
		}
	}

	for _, p := range s.Config.Ports {
		if p.Number == 0 {
			return fmt.Errorf("the zero port is reserved")
		}
		if p.Protocol.Protocol == 0 {
			return fmt.Errorf("unknown protocol for port %d", p.Number)
		}
	}

	return nil
}

// validateName asserts that a name is a valid Linux user or group name
// according to BusyBox and shadow-utils conventions. However, it disallows the
// '$' character.
func validateName(name string) error {
	if name == "" {
		return fmt.Errorf("blank name")
	}

	if name == "root" {
		return fmt.Errorf("name 'root' is reserved")
	}

	if len(name) > 32 {
		return fmt.Errorf("name is longer than 32 characters")
	}

	if reNotPOSIXPortableCharacter.MatchString(name) {
		return fmt.Errorf("name contains one or more characters not in the POSIX portable character set")
	}

	if reDigits.MatchString(name) {
		return fmt.Errorf("name comprises only digits")
	}

	if reSpecialPrefixOrSuffix.MatchString(name) {
		return fmt.Errorf("name starts or ends with a special character")
	}

	return nil
}

// BaseImage holds the components of the reference to the base image, plus
// preprocessing instructions for the base image.
type BaseImage struct {
	Repository string
	Tag        string
	Clear      Clear
}

// Reference returns the repository:tag string representation of the
// reference to the base image.
func (i BaseImage) Reference() string {
	return fmt.Sprintf("%s:%s", i.Repository, i.Tag)
}

// Clear holds toggles for clearing configuration inherited from the base
// image.
type Clear struct {
	// Clear all annotations
	Annotations bool

	// Clear the author
	Author bool

	// Clear the command
	Command bool `toml:"cmd"`

	// Unset all environment variables
	Environment bool `toml:"env"`

	// Clear the entrypoint
	Entrypoint bool `toml:"ep"`

	// Clear all labels
	Labels bool

	// Close all exposed ports
	Ports bool
}

// Packages contains instructions for the distro's canonical package manager.
type Packages struct {
	// Whether to upgrade pre-installed packages
	Upgrade bool

	// List of packages to install
	Install []string

	// Whether to clean package caches after upgrading or installing packages
	Clean bool
}

// User holds information about a Linux user.
type User struct {
	// Human-readable identifier
	Name string

	// Linux user ID (UID)
	//
	// The default value of 0 tells the program to delegate the choice of UID
	// to the user-space utility responsible for user creation.
	//
	// If not 0, then it must be an integer between 1000 and 60000, inclusive.
	ID uint `toml:"id"`

	// Whether to create a user group
	UserGroup bool `toml:"user-group"`

	// Groups to which to add the user
	Groups []string

	// GECOS field text commonly used to store a full display name
	Comment *string

	// Create a home directory for the user in /home
	CreateHome bool `toml:"create-home"`

	// Preferred interactive shell; must be a PATH-resolvable executable
	Shell string
}

// Copy holds instructions and options for copying one or more files from the
// host's file system to the working container's file system.
type Copy struct {
	// Context directory for the files to copy over to the working container
	Base string

	// Absolute path to the destination on the working container's file system
	Destination string `toml:"dest"`

	// Absolute or relative paths to source files on the host's file system;
	// may contain gitignore-style glob patterns
	Sources []string `toml:"srcs"`

	// Source files in the base directory to exclude from the copy operation;
	// may contain gitignore-style glob patterns
	Excludes []string

	// Set the mode of the copied files to this integer
	Mode uint32

	// Transfer ownership of the copied files to this user
	Owner string

	// Remove all SUID and SGID bits from the files copied to the working container
	RemoveS bool `toml:"remove-s"`
}

// Security holds security-related options for the working container.
type Security struct {
	// Options for handling files with a SUID/SGID bit
	SpecialFiles SpecialFiles `toml:"special-files"`
}

// SpecialFiles holds options for handling SUID/SGID bits.
type SpecialFiles struct {
	// Whether to remove all SUID/SGID bits automatically
	RemoveS bool `toml:"remove-s"`

	// Whether to preserve the SUID/SGID bit on one or more files
	Excludes []string
}

// Configuration holds configuration options for the image to be built from the
// working container, as defined in the OCI v1 Image Format specification.
type Configuration struct {
	// Set or update one or more annotations
	Annotations map[string]string

	// Provide contact information for the image maintainer
	Author string

	// Set the default command (or the parameters, if an entrypoint is set)
	Command []string `toml:"cmd"`

	// Describe how the image was built
	CreatedBy string `toml:"created-by"`

	// Set the entrypoint
	Entrypoint []string `toml:"ep"`

	// Set or update one or more environment variables
	Environment map[string]string `toml:"env"`

	// Set or update one or more labels
	Labels map[string]string

	// Expose one or more network ports
	Ports []Port

	// Set the default directory in which the entrypoint or command should run
	WorkDir string `toml:"work-dir"`
}

// Port holds a combination of a port number and network protocol.
type Port struct {
	Number   uint16
	Protocol ProtocolWrapper
}

// String returns a string representation of the port.
func (p Port) String() string {
	return fmt.Sprintf("%d/%s", p.Number, p.Protocol.String())
}

// Protocol is a unique identifier for a network protocol. The zero value
// represents an unknown protocol.
type Protocol uint

const (
	TCP Protocol = 1 << iota
	UDP
)

// String returns a string containing the stylized name of the protocol.
func (p Protocol) String() string {
	var s string
	switch p {
	case TCP:
		s = "tcp"
	case UDP:
		s = "udp"
	default:
		s = "unknown"
	}
	return s
}

// ProtocolWrapper wraps Protocol to facilitate its parsing from serialized data.
type ProtocolWrapper struct {
	Protocol
}

// UnmarshalText decodes the protocol from a string.
func (w *ProtocolWrapper) UnmarshalText(text []byte) error {
	var err error
	w.Protocol, err = parseProtocolString(string(text))
	return err
}

func parseProtocolString(s string) (Protocol, error) {
	var p Protocol
	switch strings.ToLower(s) {
	case "tcp":
		p = TCP
	case "udp":
		p = UDP
	default:
		return 0, fmt.Errorf("unsupported protocol: %q", s)
	}
	return p, nil
}

// Backends holds the choices of implementations of operations in the working
// container
type Backends struct {
	// Identity of the package manager in the working container
	Package pckg.ManagerWrapper

	// Identity of user-space utility for managing users and groups in the
	// working container
	User usrgrp.ManagerWrapper

	// Identity of the implementation of the find utility in the working
	// container
	Finder find.FinderWrapper
}
