// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/ok-ryoko/turret/pkg/container"
	"github.com/ok-ryoko/turret/pkg/linux"
	"github.com/ok-ryoko/turret/pkg/linux/find"
	"github.com/ok-ryoko/turret/pkg/linux/pckg"
	"github.com/ok-ryoko/turret/pkg/linux/usrgrp"
	"github.com/ok-ryoko/turret/pkg/spec"

	"github.com/containers/buildah"
	is "github.com/containers/image/v5/storage"
	"github.com/containers/storage"
	"github.com/containers/storage/pkg/archive"
	"github.com/sirupsen/logrus"
)

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

// Commit commits an image from the working container to storage, asserting
// that `repository` and `tag` are nonempty strings, and returns the ID of the
// newly created image.
func (b *Builder) Commit(
	ctx context.Context,
	store storage.Store,
	repository string,
	tag string,
	options CommitOptions,
) (string, error) {
	if repository == "" || tag == "" {
		return "", fmt.Errorf("missing image reference component")
	}

	co := buildah.CommitOptions{
		Compression:      archive.Gzip,
		HistoryTimestamp: &time.Time{},
		OmitHistory:      false,
		Squash:           true,
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
		return "", fmt.Errorf("parsing image reference: %w", err)
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

// Configure sets metadata and runtime parameters for the working container.
func (b *Builder) Configure(options ConfigureOptions) {
	b.Builder.SetOS("linux")

	if options.User != nil {
		if options.User.LoginShell != "" {
			b.Builder.SetCmd([]string{options.User.LoginShell})
		}
		b.Builder.SetUser(options.User.Name)
		b.Builder.SetWorkDir(filepath.Join("/home", options.User.Name))
	}

	for k, v := range options.Env {
		b.Builder.SetEnv(k, v)
	}

	for k, v := range options.Annotations {
		b.Builder.SetAnnotation(k, v)
	}
}

type ConfigureOptions struct {
	Annotations map[string]string
	Env         map[string]string
	User        *spec.User
}

// CopyFiles copies one or more files from the end user's home directory to the
// working container's file system, assuming `destSourcesMap` is a nonempty map
// of destinations in the working container to sources on the host. Sources are
// resolved with respect to the end user's home directory on the host;
// destinations are absolute paths in the working container's filesystem.
func (b *Builder) CopyFiles(destSourcesMap map[string][]string, options CopyFilesOptions) error {
	if len(destSourcesMap) == 0 {
		return nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting current user's home directory: %w", err)
	}

	for dest, srcs := range destSourcesMap {
		for i, s := range srcs {
			srcs[i] = fmt.Sprintf("!%s", s)
		}
		excludes := append([]string{"*"}, srcs...)
		ao := buildah.AddAndCopyOptions{
			Chown:          options.UserName,
			ContextDir:     home,
			Excludes:       excludes,
			StripSetgidBit: true,
			StripSetuidBit: true,
		}
		if err := b.Builder.Add(dest, false, ao, home); err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	return nil
}

type CopyFilesOptions struct {
	UserName string
}

// CreateUser creates the sole unprivileged user of the working container,
// asserting that `name` is a nonempty string.
func (b *Builder) CreateUser(name string, options usrgrp.CreateUserOptions) error {
	if name == "" {
		return fmt.Errorf("blank user name")
	}

	if options.ID != 0 && (options.ID < 1000 || options.ID > 60000) {
		return fmt.Errorf("UID %d outside allowed range [1000-60000]", options.ID)
	}

	if options.LoginShell != "" {
		shell, err := b.ResolveExecutable(options.LoginShell)
		if err != nil {
			return fmt.Errorf("resolving shell: %w", err)
		}
		options.LoginShell = shell
	}

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
		textOut, textErr, err := b.Run(cmd, ro)
		if err != nil {
			errMsg := "searching for special files"
			if textErr != "" {
				errMsg = fmt.Sprintf("%s (%s)", errMsg, textErr)
			}
			return fmt.Errorf("%s: %w", errMsg, err)
		}
		if len(textOut) > 0 {
			reNewline := regexp.MustCompile(`\r?\n`)
			targets = reNewline.Split(strings.TrimSpace(textOut), -1)
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

		_, textErr, err := b.Run(cmd, ro)
		if err != nil {
			errMsg := "unsetting special bit"
			if textErr != "" {
				errMsg = fmt.Sprintf("%s (%s)", errMsg, textErr)
			}
			return fmt.Errorf("%s: %w", errMsg, err)
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
	logger.Debugf("created working container from image '%s'", imageRef)

	cntr := container.Container{
		Builder: b,
		Logger:  logger,
	}

	pm, err := container.NewPackageManager(packageManager)
	if err != nil {
		return Builder{}, fmt.Errorf("creating package manager: %w", err)
	}

	um, err := container.NewUserGroupManager(userManager)
	if err != nil {
		return Builder{}, fmt.Errorf("creating user and group manager: %w", err)
	}

	f, err := find.NewCommandFactory(finder)
	if err != nil {
		return Builder{}, fmt.Errorf("creating finder: %w", err)
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
