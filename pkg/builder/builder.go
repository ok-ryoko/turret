// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package builder

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/ok-ryoko/turret/pkg/linux"
	"github.com/ok-ryoko/turret/pkg/linux/pckg"

	"github.com/containers/buildah"
	is "github.com/containers/image/v5/storage"
	"github.com/containers/storage"
	"github.com/containers/storage/pkg/archive"
	"github.com/sirupsen/logrus"
)

// TurretBuilderInterface is the interface implemented by a TurretBuilder for
// a particular Linux-based distro.
type TurretBuilderInterface interface {
	// CleanPackageCaches cleans the package caches in the working container.
	CleanPackageCaches() error

	// Commit commits an image from the working container to storage and returns
	// the ID of the newly created image.
	Commit(
		ctx context.Context,
		store storage.Store,
		repository string,
		tag string,
		options CommitOptions,
	) (string, error)

	// Configure sets runtime properties and metadata for the working container.
	Configure(user bool, options ConfigureOptions)

	// ContainerID returns the ID of the working container.
	ContainerID() string

	// CopyFiles copies files from the end user's home directory to the working
	// container's file system.
	CopyFiles(destSourcesMap map[string][]string, options CopyFilesOptions) error

	// CreateUser creates the sole unprivileged user of the working container.
	CreateUser(name string, distro linux.Distro, options CreateUserOptions) error

	// InstallPackages installs one or more packages in the working container.
	InstallPackages(packages []string) error

	// Remove removes the working container and destroys this builder, which
	// should not be used afterwards.
	Remove() error

	// UnsetSpecialBits removes the SUID/SGID bit from files in real filesystems
	// mounted in the working container.
	UnsetSpecialBits(files []string) error

	// UpgradePackages upgrades the packages in the working container.
	UpgradePackages() error
}

// TurretBuilder provides a high-level front end for Buildah for configuring
// and building container images of diverse Linux-based distros.
type TurretBuilder struct {
	// Pointer to the underlying Buildah builder instance
	Builder *buildah.Builder

	// Pointer to an object that manages packages in the working container
	PackageManager TurretPackageManagerInterface

	// Common options available to all build steps
	CommonOptions CommonOptions

	// Pointer to the underlying logger
	Logger *logrus.Logger
}

// CleanPackageCaches cleans the package caches in the working container.
func (b *TurretBuilder) CleanPackageCaches() error {
	if err := b.PackageManager.CleanCaches(b); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

// CommonOptions holds common options for every step of a build.
type CommonOptions struct {
	// Environment variables to set when running a command in the working
	// container, represented as a slice of "KEY=VALUE"s
	Env []string

	// Whether to log the output and error streams of container processes
	LogCommands bool
}

// Commit commits an image from the working container to storage, asserting
// that `repository` and `tag` are nonempty strings, and returns the ID of the
// newly created image.
func (b *TurretBuilder) Commit(
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
func (b *TurretBuilder) Configure(user bool, options ConfigureOptions) {
	b.Builder.SetOS("linux")

	if user {
		b.Builder.SetEntrypoint([]string{"/bin/sh", "-c"})
		b.Builder.SetCmd([]string{options.LoginShell})
		b.Builder.SetUser(options.UserName)
		b.Builder.SetWorkDir(filepath.Join("/home", options.UserName))
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
	LoginShell  string
	UserName    string
}

// ContainerID returns the ID of the working container.
func (b *TurretBuilder) ContainerID() string {
	return buildah.GetBuildInfo(b.Builder).ContainerID
}

// CopyFiles copies one or more files from the end user's home directory to the
// working container's file system, assuming `destSourcesMap` is a nonempty map
// of destinations in the working container to sources on the host. Sources are
// resolved with respect to the end user's home directory on the host;
// destinations are absolute paths in the working container's filesystem.
func (b *TurretBuilder) CopyFiles(destSourcesMap map[string][]string, options CopyFilesOptions) error {
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
func (b *TurretBuilder) CreateUser(name string, distro linux.Distro, options CreateUserOptions) error {
	if name == "" {
		return fmt.Errorf("blank user name")
	}

	useraddCmd := []string{"useradd", "--create-home"}

	if options.LoginShell != distro.DefaultShell() {
		shell, err := b.resolveExecutable(options.LoginShell, distro)
		if err != nil {
			return fmt.Errorf("resolving login shell: %w", err)
		}
		options.LoginShell = shell
	}
	useraddCmd = append(useraddCmd, "--shell", options.LoginShell)

	if options.ID != 0 {
		if options.ID < 1000 || options.ID > 60000 {
			return fmt.Errorf("UID %d outside allowed range [1000-60000]", options.ID)
		}
		useraddCmd = append(
			useraddCmd,
			"--uid",
			fmt.Sprintf("%d", options.ID),
		)
	}

	userGroupFlag := "--no-user-group"
	if options.UserGroup {
		userGroupFlag = "--user-group"
	}
	useraddCmd = append(useraddCmd, userGroupFlag)

	if options.Comment != "" {
		useraddCmd = append(useraddCmd, "--comment", options.Comment)
	}

	if len(options.Groups) > 0 {
		useraddCmd = append(
			useraddCmd,
			"--groups",
			strings.Join(options.Groups, ","),
		)
	}

	useraddCmd = append(useraddCmd, name)

	uao := b.defaultRunOptions()

	// CAP_DAC_READ_SEARCH and CAP_FSETID are elements of the useradd effective
	// capability set but are not needed for the operation to succeed.
	//
	uao.AddCapabilities = []string{
		"CAP_CHOWN",
		//
		// - Change owner of files copied from /etc/skel to /home/user
		// - Change owner of /var/spool/mail/user

		"CAP_DAC_OVERRIDE",
		//
		// - Open /etc/shadow and /etc/gshadow
		// - Open files copied from /etc/skel to /home/user

		"CAP_FOWNER",
		//
		// - Change owner and mode of temporary files when updating the passwd,
		// shadow, gshadow, group, subuid and subgid files in /etc
		// - Change owner and mode of /home/user and /var/spool/mail/user
		// - Change owner of, set extended attributes on and update timestamps
		// of files copied from /etc/skel to /home/user
	}

	// If the sss_cache command is available, then useradd will fork into
	// sss_cache to invalidate the System Security Services Daemon cache,
	// an operation that requires additional capabilities.
	//
	_, err := b.resolveExecutable("sss_cache", distro)
	if err != nil {
		b.Logger.Debugln("sss_cache not found; skipping cache invalidation")
	} else {
		uao.AddCapabilities = append(
			uao.AddCapabilities,
			"CAP_SETGID",
			//
			// Set the effective GID to 0 (root)

			"CAP_SETUID",
			//
			// Set the effective UID to 0 (root)
		)
	}

	if err := b.run(useraddCmd, uao); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

type CreateUserOptions struct {
	ID         uint
	UserGroup  bool
	Groups     []string
	Comment    string
	LoginShell string
}

func (b *TurretBuilder) defaultRunOptions() buildah.RunOptions {
	ro := buildah.RunOptions{
		ConfigureNetwork: buildah.NetworkDisabled,
		Quiet:            true,
	}

	if len(b.CommonOptions.Env) > 0 {
		ro.Env = append(ro.Env, b.CommonOptions.Env...)
	}

	if b.CommonOptions.LogCommands {
		ro.Logger = b.Logger
		ro.Quiet = false
	}

	return ro
}

// InstallPackages installs one or more packages to the working container.
func (b *TurretBuilder) InstallPackages(packages []string) error {
	if err := b.PackageManager.Install(b, packages); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

// Remove removes the working container and destroys this TurretBuilder, which
// should not be used afterwards.
func (b *TurretBuilder) Remove() error {
	err := b.Builder.Delete()
	if err != nil {
		return fmt.Errorf("deleting container: %w", err)
	}

	b.Builder = nil
	b.Logger = nil
	b.CommonOptions = CommonOptions{}

	return nil
}

// resolveExecutable returns the absolute path of an executable in the working
// container if it can be found and an error otherwise, assuming `command` can
// be resolved.
func (b *TurretBuilder) resolveExecutable(executable string, distro linux.Distro) (string, error) {
	shell := distro.DefaultShell()
	cmd := []string{shell}
	if filepath.Base(shell) == "bash" {
		cmd = append(cmd, "--restricted")
	}
	cmd = append(cmd, "-c", "command", "-v", executable)

	var buf bytes.Buffer
	ro := b.defaultRunOptions()
	ro.Stdout = &buf

	if err := b.run(cmd, ro); err != nil {
		return "", fmt.Errorf("running default shell or resolving executable '%s'", executable)
	}

	return strings.TrimSpace(buf.String()), nil
}

// run runs a command in the working container, optionally sanitizing and
// logging the process's standard output and error streams. When sanitizing, it
// strips all ANSI escape codes as well as superfluous whitespace.
func (b *TurretBuilder) run(cmd []string, options buildah.RunOptions) error {
	var stderrBuf bytes.Buffer
	if options.Stderr == nil && b.CommonOptions.LogCommands {
		options.Stderr = &stderrBuf
	}

	var stdoutBuf bytes.Buffer
	if options.Stdout == nil && b.CommonOptions.LogCommands {
		options.Stdout = &stdoutBuf
	}

	defer func() {
		if b.CommonOptions.LogCommands {
			reEscape := regexp.MustCompile(`((\\x1b|\\u001b)\[[0-9;]*[A-Za-z]?)+`)
			reWhitespace := regexp.MustCompile(`[[:space:]]+`)

			if stderrBuf.Len() > 0 {
				lines := stderrBuf.String()
				for _, l := range strings.Split(lines, "\n") {
					l = strings.Map(func(r rune) rune {
						if unicode.IsGraphic(r) {
							return r
						}
						return -1
					}, l)
					l = reWhitespace.ReplaceAllLiteralString(strings.TrimSpace(l), " ")
					if l == "" {
						continue
					}
					l = reEscape.ReplaceAllLiteralString(l, "")
					b.Logger.Debugf("%s: stderr: %s", cmd[0], l)
				}
			}

			if stdoutBuf.Len() > 0 {
				lines := stdoutBuf.String()
				for _, l := range strings.Split(lines, "\n") {
					l = strings.Map(func(r rune) rune {
						if unicode.IsGraphic(r) {
							return r
						}
						return -1
					}, l)
					l = reWhitespace.ReplaceAllLiteralString(strings.TrimSpace(l), " ")
					if l == "" {
						continue
					}
					l = reEscape.ReplaceAllLiteralString(l, "")
					b.Logger.Debugf("%s: stdout: %s", cmd[0], l)
				}
			}
		}
	}()

	if err := b.Builder.Run(cmd, options); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

// UnsetSpecialBits removes the SUID/SGID bit from files in the working container,
// assuming the availability of the chmod(1) and find(1) core utilities and
// searching only real file systems, excluding the /home directory.
func (b *TurretBuilder) UnsetSpecialBits(excludes []string) error {
	findCmd := []string{
		"find", "/",
		"-xdev",
		"!", "(", "-wholename", "/home", "-prune", ")",
		"-perm", "/u=s,g=s",
	}

	var buf bytes.Buffer
	fo := b.defaultRunOptions()
	fo.Stdout = &buf

	if err := b.run(findCmd, fo); err != nil {
		return fmt.Errorf("%w", err)
	}

	targets := strings.Split(strings.TrimSpace(buf.String()), "\n")
	for i, t := range targets {
		targets[i] = strings.TrimSpace(t)
	}

	if len(excludes) > 0 {
		excludeSet := map[string]bool{}
		for _, e := range excludes {
			excludeSet[e] = true
		}

		var filteredTargets []string
		for _, t := range targets {
			if _, ok := excludeSet[t]; ok {
				continue
			}
			filteredTargets = append(filteredTargets, t)
		}

		targets = filteredTargets
	}

	chmodCmd := []string{"chmod", "-s"}
	chmodCmd = append(chmodCmd, targets...)

	co := b.defaultRunOptions()

	if err := b.run(chmodCmd, co); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

// UpgradePackages upgrades the packages in the working container.
func (b *TurretBuilder) UpgradePackages() error {
	if err := b.PackageManager.Upgrade(b); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

// New creates a new TurretBuilder for a given combination of a Linux-based
// distro and package manager.
func New(
	ctx context.Context,
	distro linux.Distro,
	packageManager pckg.Manager,
	imageRef string,
	pull bool,
	store storage.Store,
	logger *logrus.Logger,
	options CommonOptions,
) (TurretBuilderInterface, error) {
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
		return nil, fmt.Errorf("creating Buildah builder: %w", err)
	}
	logger.Debugf("created working container from image '%s'", imageRef)

	pm, err := NewPackageManager(packageManager)
	if err != nil {
		return nil, fmt.Errorf("creating package manager: %w", err)
	}

	var tb TurretBuilderInterface
	switch distro {
	case linux.Alpine:
		tb = &AlpineTurretBuilder{
			TurretBuilder{
				Builder:        b,
				PackageManager: pm,
				Logger:         logger,
				CommonOptions:  options,
			},
		}
	case linux.Debian:
		options.Env = append(options.Env, "DEBIAN_FRONTEND=noninteractive")
		tb = &TurretBuilder{
			Builder:        b,
			PackageManager: pm,
			Logger:         logger,
			CommonOptions:  options,
		}
	case
		linux.Arch,
		linux.Chimera,
		linux.Fedora,
		linux.OpenSUSE,
		linux.Void:
		tb = &TurretBuilder{
			Builder:        b,
			PackageManager: pm,
			Logger:         logger,
			CommonOptions:  options,
		}
	default:
		return nil, fmt.Errorf("unrecognized distro")
	}
	return tb, nil
}
