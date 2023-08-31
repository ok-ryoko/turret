// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/ok-ryoko/turret/pkg/container"
	"github.com/ok-ryoko/turret/pkg/linux"
	"github.com/ok-ryoko/turret/pkg/linux/find"
	"github.com/ok-ryoko/turret/pkg/linux/pckg"
	"github.com/ok-ryoko/turret/pkg/linux/usrgrp"

	"github.com/containers/buildah"
	is "github.com/containers/image/v5/storage"
	"github.com/containers/storage"
	"github.com/containers/storage/pkg/archive"
	"github.com/sirupsen/logrus"
)

const manifestType string = "application/vnd.oci.image.manifest.v1+json"

// Builder provides a high-level front end for Buildah for configuring and
// building container images of diverse Linux-based distros.
type Builder struct {
	container.Container

	// Pointer to an object that manages packages in the working container
	PackageManager container.PackageManagerInterface

	// Pointer to an object that manages users and groups in the working
	// container
	UserGroupManager container.UserGroupManagerInterface

	// Object that creates commands for locating files in the working container
	FinderCommandFactory find.CommandFactory
}

// CleanPackageCaches cleans the package caches in the working container.
func (b *Builder) CleanPackageCaches() error {
	if err := b.PackageManager.CleanCaches(&b.Container); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

// Commit commits an image from the working container to storage and returns
// the ID of the newly created image, assuming `repository` and `tag` are
// nonempty strings from which a valid image reference can be composed.
func (b *Builder) Commit(
	ctx context.Context,
	store storage.Store,
	repository string,
	tag string,
	options CommitOptions,
) (string, error) {
	co := buildah.CommitOptions{
		PreferredManifestType: manifestType,
		Compression:           archive.Gzip,
		HistoryTimestamp:      &time.Time{},
		OmitHistory:           false,
		Squash:                true,
	}

	if options.Latest && tag != "latest" {
		co.AdditionalTags = append(
			co.AdditionalTags,
			fmt.Sprintf("%s:latest", repository),
		)
	}

	if options.KeepHistory {
		co.HistoryTimestamp = nil
		co.OmitHistory = false
	}

	imageRef := fmt.Sprintf("%s:%s", repository, tag)
	storageRef, err := is.Transport.ParseStoreReference(store, imageRef)
	if err != nil {
		return "", fmt.Errorf("parsing reference: %w", err)
	}
	imageID, _, _, err := b.Builder.Commit(ctx, storageRef, co)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}
	return imageID, nil
}

type CommitOptions struct {
	KeepHistory bool
	Latest      bool
}

// Configure alters the metadata on and execution of the working container.
func (b *Builder) Configure(options ConfigureOptions) {
	if options.ClearAnnotations {
		for k := range b.Builder.Annotations() {
			if !strings.HasPrefix("org.opencontainers.image.base", k) {
				b.Builder.UnsetAnnotation(k)
			}
		}
	}
	for k, v := range options.Annotations {
		b.Builder.SetAnnotation(k, v)
	}

	if options.ClearAuthor {
		b.Builder.SetMaintainer("")
	}
	if options.Author != "" {
		b.Builder.SetMaintainer(options.Author)
	}

	if options.ClearCommand {
		b.Builder.SetCmd([]string{})
	}
	if len(options.Command) > 0 {
		b.Builder.SetCmd(options.Command)
	}

	if options.CreatedBy != "" {
		b.Builder.SetCreatedBy(options.CreatedBy)
	}

	if options.ClearEntrypoint {
		b.Builder.SetEntrypoint([]string{})
	}
	if len(options.Entrypoint) > 0 {
		b.Builder.SetEntrypoint(options.Entrypoint)
	}

	if options.ClearEnvironment {
		b.Builder.ClearEnv()
	}
	for k, v := range options.Environment {
		b.Builder.SetEnv(k, v)
	}

	if options.ClearLabels {
		b.Builder.ClearLabels()
	}
	for k, v := range options.Labels {
		b.Builder.SetLabel(k, v)
	}

	b.Builder.SetOS("linux")

	if options.ClearPorts {
		b.Builder.ClearPorts()
	}
	if len(options.Ports) > 0 {
		for _, p := range options.Ports {
			b.Builder.SetPort(p)
		}
	}

	if options.WorkDir != "" {
		b.Builder.SetWorkDir(options.WorkDir)
	}

	if options.User != nil {
		b.Builder.SetUser(options.User.Name)
		if options.WorkDir == "" && options.User.CreateHome {
			b.Builder.SetWorkDir(filepath.Join("/home", options.User.Name))
		}
	}
}

// ConfigureOptions holds configuration options for the working container.
type ConfigureOptions struct {
	// Clear all annotations inherited from the base image
	ClearAnnotations bool

	// Set or update one or more annotations
	Annotations map[string]string

	// Clear the author inherited from the base image
	ClearAuthor bool

	// Provide contact information for the image maintainer
	Author string

	// Clear the command inherited from the base image
	ClearCommand bool

	// Set the default command (or the parameters, if an entrypoint is set)
	Command []string

	// Describe how the image was built
	CreatedBy string

	// Clear the entrypoint inherited from the base image
	ClearEntrypoint bool

	// Set the entrypoint
	Entrypoint []string

	// Unset all environment variables inherited from the base image
	ClearEnvironment bool

	// Set or update one or more environment variables
	Environment map[string]string

	// Clear all labels inherited from the base image
	ClearLabels bool

	// Set or update one or more labels
	Labels map[string]string

	// Close all exposed ports inherited from the base image
	ClearPorts bool

	// Expose one or more network ports
	Ports []string

	// Set the user as whom the entrypoint or command should run
	User *ConfigureUserOptions

	// Set the default directory in which the entrypoint or command should run
	WorkDir string
}

type ConfigureUserOptions struct {
	Name       string
	CreateHome bool
}

// CopyFiles copies one or more files on the host's file system to the working
// container's file system, assuming `base` and `dest` are absolute file paths
// and `srcs` is a nonempty slice of file paths.
//
// `base` is an absolute path to a directory on the host's file system against
// which relative paths in `srcs` should be resolved.
//
// `dest` is an absolute path to a destination on the working container's file
// system. If the destination ends with a path separator, then it's assumed to
// be a directory.
//
// `srcs` is a slice of relative or absolute paths to items on the host's file
// system. Relative paths are resolved with respect to `base`.
//
// If there is only one source item and the destination does not end with a
// path separator, then copy the item to the parent directory in the
// destination, renaming the item to match the destination as needed.
func (b *Builder) CopyFiles(base string, dest string, srcs []string, options CopyFilesOptions) error {
	patterns := make([]string, len(srcs))
	for i, s := range srcs {
		patterns[i] = fmt.Sprintf("!%s", s)
	}
	excludes := append([]string{"*"}, patterns...)
	if len(options.Excludes) > 0 {
		excludes = append(excludes, options.Excludes...)
	}

	aco := buildah.AddAndCopyOptions{
		ContextDir: base,
		Excludes:   excludes,
	}

	if options.Owner != "" {
		aco.Chown = options.Owner
	}

	if options.Mode != 0 {
		aco.Chmod = fmt.Sprint(options.Mode)
	}

	if options.RemoveS {
		aco.StripSetuidBit = true
		aco.StripSetgidBit = true
	}

	if err := b.Builder.Add(dest, false, aco, base); err != nil {
		return fmt.Errorf("copying files from %q to %q: %w", base, dest, err)
	}

	return nil
}

// CopyFilesOptions holds options for copying files from the host's file system
// to the working container's file system.
type CopyFilesOptions struct {
	// Source files in the base directory to exclude from the copy operation;
	// may contain gitignore-style glob patterns
	Excludes []string

	// Set the mode of the copied files to this integer
	Mode uint32

	// Transfer ownership of the copied files to this user
	Owner string

	// Remove all SUID and SGID bits from the files copied to the working
	// container
	RemoveS bool
}

// CreateUser creates the sole unprivileged user of the working container,
// assuming `name` is a nonempty string.
func (b *Builder) CreateUser(name string, options usrgrp.CreateUserOptions) error {
	if err := b.UserGroupManager.CreateUser(&b.Container, name, options); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

// InstallPackages installs one or more packages to the working container.
func (b *Builder) InstallPackages(packages []string) error {
	if err := b.PackageManager.Install(&b.Container, packages); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

// UnsetSpecialBits removes the SUID/SGID bit from files in the working
// container, assuming the availability of the chmod and find core
// utilities and searching only real (non-device) file systems.
//
// `excludes` is a slice of absolute paths to real files in the working
// container for which to keep the SUID/SGID bit.
func (b *Builder) UnsetSpecialBits(excludes []string) error {
	var targets []string

	{
		cmd, capabilities := b.FinderCommandFactory.NewFindSpecialCmd()
		ro := b.DefaultRunOptions()
		ro.AddCapabilities = capabilities
		outText, errText, err := b.Run(cmd, ro)
		if err != nil {
			errContext := "searching for special files"
			if errText != "" {
				errContext = fmt.Sprintf("%s (%q)", errContext, errText)
			}
			return fmt.Errorf("%s: %w", errContext, err)
		}
		if len(outText) > 0 {
			targets = strings.Split(strings.ReplaceAll(strings.TrimSpace(outText), "\r\n", "\n"), "\n")
		}
	}

	if len(excludes) > 0 {
		excludeSet := map[string]bool{}
		for _, e := range excludes {
			excludeSet[e] = true
		}

		var filteredTargets []string
		for _, t := range targets {
			if _, ok := excludeSet[t]; !ok {
				filteredTargets = append(filteredTargets, t)
			}
		}

		targets = filteredTargets
	}

	if len(targets) > 0 {
		cmd := append([]string{"chmod", "-s"}, targets...)

		// CAP_FSETID is a member of the chmod effective capability set but is
		// neither sufficient nor necessary for this operation
		//
		ro := b.DefaultRunOptions()
		ro.AddCapabilities = []string{
			"CAP_DAC_READ_SEARCH",
			"CAP_FOWNER",
		}

		_, errText, err := b.Run(cmd, ro)
		if err != nil {
			errContext := "unsetting special bit"
			if errText != "" {
				errContext = fmt.Sprintf("%s (%q)", errContext, errText)
			}
			return fmt.Errorf("%s: %w", errContext, err)
		}
	}

	return nil
}

// UpgradePackages upgrades the packages in the working container.
func (b *Builder) UpgradePackages() error {
	if err := b.PackageManager.Upgrade(&b.Container); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

// NewBuilder creates a new Builder for a given combination of a Linux-
// based distro, package manager, and user/group management utility.
func New(
	ctx context.Context,
	distro linux.Distro,
	packageManager pckg.Manager,
	userManager usrgrp.Manager,
	finder find.Finder,
	imageRef string,
	pull bool,
	store storage.Store,
	logger *logrus.Logger,
	options container.CommonOptions,
) (Builder, error) {
	bo := buildah.BuilderOptions{
		Capabilities: []string{},
		FromImage:    imageRef,
		Isolation:    buildah.IsolationOCIRootless,
		PullPolicy:   buildah.PullNever,
	}
	if pull {
		bo.PullPolicy = buildah.PullIfMissing
	}
	if options.LogCommands {
		bo.Logger = logger
	}

	b, err := buildah.NewBuilder(ctx, store, bo)
	if err != nil {
		return Builder{}, fmt.Errorf("creating Buildah builder: %w", err)
	}
	logger.Debugf("created working container from image %s", imageRef)

	cntr := container.Container{
		Builder: b,
		Logger:  logger,
	}

	pm, err := container.NewPackageManager(packageManager)
	if err != nil {
		return Builder{}, fmt.Errorf("creating package management interface: %w", err)
	}

	um, err := container.NewUserGroupManager(userManager)
	if err != nil {
		return Builder{}, fmt.Errorf("creating user management interface: %w", err)
	}

	f, err := find.NewCommandFactory(finder)
	if err != nil {
		return Builder{}, fmt.Errorf("creating find command factory: %w", err)
	}

	if distro == linux.Debian {
		options.Env = append(options.Env, "DEBIAN_FRONTEND=noninteractive")
	}
	cntr.CommonOptions = options

	return Builder{
		Container:            cntr,
		PackageManager:       pm,
		UserGroupManager:     um,
		FinderCommandFactory: f,
	}, nil
}
