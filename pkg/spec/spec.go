// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/ok-ryoko/turret/pkg/linux"
	"github.com/ok-ryoko/turret/pkg/linux/find"
	"github.com/ok-ryoko/turret/pkg/linux/pckg"
	"github.com/ok-ryoko/turret/pkg/linux/usrgrp"
)

const (
	rePOSIXPortableName    string = `^[0-9A-Za-z]$|^[0-9A-Za-z][-\._0-9A-Za-z]*[0-9A-Za-z]$`
	reInvalidName          string = `^[0-9]+$|^\.{1,2}$`
	reReverseUnlimitedFQDN string = `^\.?([0-9A-Za-z]|[0-9A-Za-z][-0-9A-Za-z]*[0-9A-Za-z]\.)*[0-9A-Za-z]$`
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

	// Map of environment variables to set or update in the working container
	Env map[string]string

	// Map of annotations to apply to the working container
	Annotations map[string]string

	// Whether to preserve the image history and timestamps of the files in the
	// working container's file system
	KeepHistory bool `toml:"keep-history"`

	// Security options for the working container
	Security Security

	// Choices of implementations of operations in the working container
	Backends Backends
}

// Fill populates empty optional fields in the spec using information encoded
// by required fields.
func (s *Spec) Fill() {
	if s.Backends.Package.Manager == 0 {
		s.Backends.Package = pckg.ManagerWrapper{
			Manager: s.Distro.DefaultPackageManager(),
		}
	}

	if s.Backends.User.Manager == 0 {
		s.Backends.User = usrgrp.ManagerWrapper{
			Manager: s.Distro.DefaultUserManager(),
		}
	}

	if s.Backends.Finder.Finder == 0 {
		s.Backends.Finder = find.FinderWrapper{
			Finder: s.Distro.DefaultFinder(),
		}
	}

	if s.Annotations == nil {
		s.Annotations = map[string]string{}
	}

	if s.Env == nil {
		s.Env = map[string]string{}
	}
}

// Validate asserts that the spec is suitable for ingestion by a builder.
func (s *Spec) Validate() error {
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
		reProper := regexp.MustCompile(rePOSIXPortableName)
		reImproper := regexp.MustCompile(reInvalidName)

		if s.User.Name != "" {
			if s.User.Name == "root" {
				return fmt.Errorf("won't create unprivileged user named 'root'")
			}

			if len(s.User.Name) > 32 {
				return fmt.Errorf("user name '%s' too long (limit: 32 chars)", s.User.Name)
			}

			if !reProper.MatchString(s.User.Name) || reImproper.MatchString(s.User.Name) {
				return fmt.Errorf("invalid user name '%s'", s.User.Name)
			}
		}

		if s.User.ID != 0 && (s.User.ID < 1000 || s.User.ID > 60000) {
			return fmt.Errorf("UID %d outside allowed range [1000-60000]", s.User.ID)
		}

		if len(s.User.Groups) > 0 {
			for _, g := range s.User.Groups {
				if !reProper.MatchString(g) || reImproper.MatchString(g) {
					return fmt.Errorf("invalid group name '%s'", g)
				}
			}
		}
	}

	if len(s.Copy) > 0 {
		for i, c := range s.Copy {
			if c.Base == "" {
				return fmt.Errorf("missing base")
			}
			if !filepath.IsAbs(c.Base) {
				return fmt.Errorf("base is not an absolute path")
			}
			s.Copy[i].Base = filepath.Clean(c.Base)

			if c.Destination == "" {
				return fmt.Errorf("missing destination")
			}
			if !filepath.IsAbs(c.Destination) {
				return fmt.Errorf("destination is not an absolute path")
			}
			s.Copy[i].Destination = filepath.Clean(c.Destination)

			if len(c.Sources) == 0 {
				return fmt.Errorf("missing sources for destination %q", c.Destination)
			}
			reScheme := regexp.MustCompile(`^[^:/?#]+:`) // Network Working Group RFC 3986 Appendix B
			for j, src := range c.Sources {
				if reScheme.MatchString(src) {
					return fmt.Errorf("only schemeless paths are supported (%q)", src)
				}
				s.Copy[i].Sources[j] = filepath.Clean(src)
			}
		}
	}

	if len(s.Annotations) > 0 {
		re := regexp.MustCompile(reReverseUnlimitedFQDN)
		for k := range s.Annotations {
			if !re.MatchString(k) {
				return fmt.Errorf("invalid annotation key '%s'", k)
			}
		}
	}

	return nil
}

// BaseImage holds the components of the reference to the base image.
type BaseImage struct {
	Repository string
	Tag        string
}

// Reference returns the repository:tag string representation of the
// reference to the base image.
func (i BaseImage) Reference() string {
	return fmt.Sprintf("%s:%s", i.Repository, i.Tag)
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

	// Login shell; must be a PATH-resolvable executable
	LoginShell string `toml:"login-shell"`
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
