// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"fmt"
	"regexp"

	"github.com/ok-ryoko/turret/pkg/linux"
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

// Fill populates empty optional fields in the spec using information encoded
// by required fields.
func (s *Spec) Fill() {
	if s.Packages.Manager.Manager == 0 {
		s.Packages.Manager = pckg.ManagerWrapper{
			Manager: s.Distro.DefaultPackageManager(),
		}
	}

	if s.User.Manager.Manager == 0 {
		s.User.Manager = usrgrp.ManagerWrapper{
			Manager: s.Distro.DefaultUserManager(),
		}
	}

	if s.Annotations == nil {
		s.Annotations = map[string]string{}
	}

	if s.Copy == nil {
		s.Copy = map[string][]string{}
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

	if s.Packages.Manager.Manager == 0 {
		return fmt.Errorf("missing package manager")
	}

	if s.Repository == "" {
		return fmt.Errorf("missing image repository (name)")
	}

	if s.From.Repository == "" || s.From.Tag == "" {
		return fmt.Errorf("missing base image repository (name) or tag")
	}

	if len(s.Packages.Install) > 0 {
		re := regexp.MustCompile(s.Packages.Manager.RePackageName())
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
	// Identity of the package manager in the working container
	Manager pckg.ManagerWrapper

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
	Comment string

	// Login shell; must be a PATH-resolvable executable
	LoginShell string `toml:"login-shell"`

	// Choice of user-space utility for managing users and groups
	Manager usrgrp.ManagerWrapper
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
