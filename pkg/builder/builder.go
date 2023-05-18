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

	"github.com/containers/buildah"
	is "github.com/containers/image/v5/storage"
	"github.com/containers/storage"
	"github.com/containers/storage/pkg/archive"
	"github.com/sirupsen/logrus"
)

// TurretBuilderInterface is the interface implemented by a Turret builder for
// a particular GNU/Linux distro
type TurretBuilderInterface interface {
	// CleanPackageCaches cleans the package cache in the working container
	// using the distro's canonical package manager
	CleanPackageCaches() error

	// Commit commits an image from the working container to storage and returns
	// the ID of the newly created image
	Commit(
		ctx context.Context,
		store storage.Store,
		repository string,
		tag string,
		options CommitOptions,
	) (string, error)

	// Configure sets runtime properties and metadata for the working container
	Configure(user bool, options ConfigureOptions)

	// ContainerID returns the ID of the working container
	ContainerID() string

	// CopyFiles copies files from the end user's home directory to the working
	// container's file system
	CopyFiles(destSourcesMap map[string][]string, options CopyFilesOptions) error

	// CreateUser creates the sole unprivileged user of the working container
	CreateUser(name string, distro GNULinuxDistro, options CreateUserOptions) error

	// Distro returns a representation of the GNU/Linux distribution for which this
	// builder is specialized
	Distro() GNULinuxDistro

	// InstallPackages installs one or more packages in the working container
	// using the distro's canonical package manager
	InstallPackages(packages []string) error

	// Remove removes the working container and destroys this builder, which should
	// not be used afterwards
	Remove() error

	// UnsetSpecialBits removes SUID and SGID bits from files in the working
	// container
	UnsetSpecialBits(files []string) error

	// UpgradePackages upgrades the packages in the working container using the
	// distro's canonical package manager
	UpgradePackages() error
}

// TurretBuilder provides a high-level API over Buildah
type TurretBuilder struct {
	// Pointer to the underlying Buildah Builder instance
	Builder *buildah.Builder

	// Common options available to all build steps
	CommonOptions CommonOptions

	// Pointer to the underlying logger
	Logger *logrus.Logger
}

// CommonOptions holds common options for every step of a build
type CommonOptions struct {
	// Whether to log the output and error streams of container processes
	LogCommands bool
}

// Commit commits an image from the working container to storage and returns
// the ID of the newly created image;
// asserts that `ref` is a nonempty string
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

	ref := fmt.Sprintf("%s:%s", repository, tag)
	storageRef, err := is.Transport.ParseStoreReference(store, ref)
	if err != nil {
		return "", fmt.Errorf("parsing image reference: %w", err)
	}
	imageId, _, _, err := b.Builder.Commit(ctx, storageRef, co)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}
	return imageId, nil
}

type CommitOptions struct {
	KeepHistory bool
	Latest      bool
}

// Configure sets runtime properties and metadata for the working container
func (b *TurretBuilder) Configure(user bool, options ConfigureOptions) {
	b.Builder.SetOS("linux")

	if user {
		b.Builder.SetUser(options.UserName)
		b.Builder.SetEntrypoint([]string{"/bin/sh", "-c"})
		b.Builder.SetCmd([]string{options.LoginShell})
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

// ContainerID returns the ID of the working container
func (b *TurretBuilder) ContainerID() string {
	return buildah.GetBuildInfo(b.Builder).ContainerID
}

// CopyFiles copies files from the end user's home directory to the working
// container's file system;
// does nothing if `copyMap` is empty
func (b *TurretBuilder) CopyFiles(destSourcesMap map[string][]string, options CopyFilesOptions) error {
	if len(destSourcesMap) == 0 {
		return nil
	}

	hostUserHomeDir, err := os.UserHomeDir()
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
			ContextDir:     hostUserHomeDir,
			Excludes:       excludes,
			StripSetgidBit: true,
			StripSetuidBit: true,
		}

		if err := b.Builder.Add(dest, false, ao, hostUserHomeDir); err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	return nil
}

type CopyFilesOptions struct {
	UserName string
}

// CreateUser creates the sole unprivileged user of the working container;
// asserts that `name` is a nonempty string
func (b *TurretBuilder) CreateUser(name string, distro GNULinuxDistro, options CreateUserOptions) error {
	if name == "" {
		return fmt.Errorf("blank user name")
	}

	if options.LoginShell != distro.DefaultShell() {
		shell, err := b.resolveExecutable(options.LoginShell, distro)
		if err != nil {
			return fmt.Errorf("resolving login shell: %w", err)
		}
		options.LoginShell = shell
	}

	useraddCmd := []string{
		"useradd",
		"--create-home",
		"--uid", fmt.Sprintf("%d", options.UID),
		"--shell", options.LoginShell,
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
	uao.AddCapabilities = []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_FOWNER",
	}

	if err := b.run(useraddCmd, uao); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

type CreateUserOptions struct {
	UID        uint
	UserGroup  bool
	Groups     []string
	Comment    string
	LoginShell string
}

func (b *TurretBuilder) defaultRunOptions() buildah.RunOptions {
	options := buildah.RunOptions{
		ConfigureNetwork: buildah.NetworkDisabled,
		Quiet:            true,
	}

	if b.CommonOptions.LogCommands {
		options.Logger = b.Logger
		options.Quiet = false
	}

	return options
}

// Remove removes the working container and destroys this builder, which should
// not be used afterwards
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
// container if it can be found and an error otherwise;
// assumes the availability of the `command` shell built-in
func (b *TurretBuilder) resolveExecutable(executable string, distro GNULinuxDistro) (string, error) {
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
// logging the process's standard output and error streams;
// strips all ANSI escape codes as well as superfluous whitespace
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
			re1 := regexp.MustCompile(`([\\x1b|\\u001b]\[[0-9;]*[A-Za-z]?)+`)
			re2 := regexp.MustCompile(`[[:space:]]+`)

			if stderrBuf.Len() > 0 {
				lines := stderrBuf.String()
				for _, l := range strings.Split(lines, "\n") {
					l = strings.Map(func(r rune) rune {
						if unicode.IsGraphic(r) {
							return r
						}
						return -1
					}, l)
					l = re2.ReplaceAllLiteralString(strings.TrimSpace(l), " ")
					if l == "" {
						continue
					}
					l = re1.ReplaceAllLiteralString(l, "")
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
					l = re2.ReplaceAllLiteralString(strings.TrimSpace(l), " ")
					if l == "" {
						continue
					}
					l = re1.ReplaceAllLiteralString(l, "")
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

// UnsetSpecialBits removes SUID and SGID bits from files and directories in
// the working container;
// assumes the availability of the chmod(1) utility;
// does nothing if `files` is empty
func (b *TurretBuilder) UnsetSpecialBits(files []string) error {
	if len(files) == 0 {
		return nil
	}

	cmd := []string{"chmod", "-s"}
	cmd = append(cmd, files...)

	ro := b.defaultRunOptions()

	if err := b.run(cmd, ro); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}

// New creates a Turret builder
func New(
	ctx context.Context,
	distro GNULinuxDistro,
	image string,
	store storage.Store,
	logger *logrus.Logger,
	options CommonOptions,
) (TurretBuilderInterface, error) {
	bo := buildah.BuilderOptions{
		Capabilities: []string{},
		FromImage:    image,
		Isolation:    buildah.IsolationOCIRootless,
		PullPolicy:   buildah.PullNever,
	}
	if options.LogCommands {
		bo.Logger = logger
	}

	b, err := buildah.NewBuilder(ctx, store, bo)
	if err != nil {
		return nil, fmt.Errorf("creating Buildah builder: %w", err)
	}
	logger.Debugf("created working container from image '%s'", image)

	var tb TurretBuilderInterface
	switch distro {
	case Alpine:
		tb = &AlpineTurretBuilder{
			TurretBuilder: TurretBuilder{
				Builder:       b,
				Logger:        logger,
				CommonOptions: options,
			},
		}
	case Arch:
		tb = &ArchTurretBuilder{
			TurretBuilder: TurretBuilder{
				Builder:       b,
				Logger:        logger,
				CommonOptions: options,
			},
		}
	case Debian:
		tb = &DebianTurretBuilder{
			TurretBuilder: TurretBuilder{
				Builder:       b,
				Logger:        logger,
				CommonOptions: options,
			},
		}
	case Fedora:
		tb = &FedoraTurretBuilder{
			TurretBuilder: TurretBuilder{
				Builder:       b,
				Logger:        logger,
				CommonOptions: options,
			},
		}
	case OpenSUSE:
		tb = &OpenSUSETurretBuilder{
			TurretBuilder: TurretBuilder{
				Builder:       b,
				Logger:        logger,
				CommonOptions: options,
			},
		}
	case Void:
		tb = &VoidTurretBuilder{
			TurretBuilder: TurretBuilder{
				Builder:       b,
				Logger:        logger,
				CommonOptions: options,
			},
		}
	default:
		return nil, fmt.Errorf("unrecognized distro")
	}
	return tb, nil
}
