// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ok-ryoko/turret/pkg/builder"
	"github.com/ok-ryoko/turret/pkg/container"
	"github.com/ok-ryoko/turret/pkg/linux/usrgrp"
	"github.com/ok-ryoko/turret/pkg/spec"

	"github.com/containers/storage"
	"github.com/containers/storage/pkg/unshare"
	"github.com/pelletier/go-toml/v2"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func newBuildCmd(logger *logrus.Logger) *cli.Command {
	return &cli.Command{
		Name:                   "build",
		Aliases:                []string{"b"},
		Usage:                  "Build an OCI image from a Turret spec",
		ArgsUsage:              "SPEC",
		HideHelpCommand:        true,
		UseShortOptionHandling: true,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "Overwrite the image if it already exists",
				Value:   false,
			},
			&cli.BoolFlag{
				Name:    "hash-spec",
				Aliases: []string{"H"},
				Usage:   "Annotate the image with the SHA256 hash of SPEC",
				Value:   false,
			},
			&cli.BoolFlag{
				Name:    "keep",
				Aliases: []string{"k"},
				Usage:   "Retain the working container",
				Value:   false,
			},
			&cli.BoolFlag{
				Name:    "latest",
				Aliases: []string{"l"},
				Usage:   "Create or update the 'latest' tag",
				Value:   false,
			},
			&cli.BoolFlag{
				Name:    "pull",
				Aliases: []string{"p"},
				Usage:   "Pull the base image from remote storage if it doesn't exist locally",
				Value:   false,
			},
			&cli.BoolFlag{
				Name:    "quiet",
				Aliases: []string{"q"},
				Usage:   "Print nothing (overriding alias for --verbosity 0)",
				Value:   false,
			},
			&cli.UintFlag{
				Name:    "verbosity",
				Aliases: []string{"v"},
				Usage:   "Set the output level, from nothing (0) to everything (4)",
				Value:   1,
			},
		},
		Action: func(cCtx *cli.Context) error {
			if !cCtx.Args().Present() {
				if err := cli.ShowCommandHelp(cCtx, cCtx.Command.Name); err != nil {
					return fmt.Errorf("displaying help: %w", err)
				}
				return nil
			}

			unshare.MaybeReexecUsingUserNamespace(true)
			ctx := context.Background()

			v := cCtx.Uint("verbosity")
			if cCtx.Bool("quiet") {
				v = 0
			}
			setLoggerLevel(logger, v)

			specPath, err := processPath(cCtx.Args().First())
			if err != nil {
				return fmt.Errorf("processing path: %w", err)
			}
			logger.Debugln("processed spec path")

			spec, digest, err := createSpec(specPath, cCtx.Bool("hash-spec"))
			if err != nil {
				return fmt.Errorf("creating build spec: %w", err)
			}
			logger.Debugln("created build spec")

			storeOptions, err := storage.DefaultStoreOptionsAutoDetectUID()
			if err != nil {
				storeOptions = storage.StoreOptions{}
			}
			store, err := storage.GetStore(storeOptions)
			if err != nil {
				return fmt.Errorf("creating store: %w", err)
			}
			defer func() {
				layers, shutdownErr := store.Shutdown(false)
				if shutdownErr != nil {
					logger.Warnln("failed releasing driver resources")
					logger.Infoln(
						"the following layers may still be mounted:",
						strings.Join(layers, ", "),
					)
				}
			}()

			imageRef := spec.Repository
			if spec.Tag != "" {
				imageRef = fmt.Sprintf("%s:%s", imageRef, spec.Tag)
			}
			localRef := fmt.Sprintf("localhost/%s", imageRef)

			if store.Exists(localRef) && !cCtx.Bool("force") {
				return fmt.Errorf("image %s already exists", localRef)
			}

			distro := spec.Distro.Distro
			packageManager := spec.Backends.Package.Manager
			userManager := spec.Backends.User.Manager
			finder := spec.Backends.Finder.Finder
			baseRef := spec.From.Reference()
			commonOptions := container.CommonOptions{LogCommands: v >= 4}
			b, err := builder.New(
				ctx,
				distro,
				packageManager,
				userManager,
				finder,
				baseRef,
				cCtx.Bool("pull"),
				store,
				logger,
				commonOptions,
			)
			if err != nil {
				return fmt.Errorf("creating %s Linux Turret builder: %w", distro.String(), err)
			}
			defer func() {
				if !cCtx.Bool("keep") {
					if removeErr := b.Remove(); removeErr != nil {
						logger.Warnln("failed deleting working container")
						logger.Infoln("please remove the container manually: buildah rm", b.ContainerID())
					}
				}
			}()
			logger.Debugf("created %s Linux Turret builder", distro.String())

			if spec.Packages.Upgrade {
				logger.Debugln("upgrading packages...")
				if err = b.UpgradePackages(); err != nil {
					return fmt.Errorf("upgrading packages: %w", err)
				}
				logger.Debugln("upgrade step succeeded")
			}

			if len(spec.Packages.Install) > 0 {
				logger.Debugln("installing packages...")
				if err = b.InstallPackages(spec.Packages.Install); err != nil {
					return fmt.Errorf("installing packages: %w", err)
				}
				logger.Debugln("package installation step succeeded")
			}

			if spec.Packages.Clean {
				if err = b.CleanPackageCaches(); err != nil {
					return fmt.Errorf("cleaning package caches: %w", err)
				}
				logger.Debugln("package cache cleaning step succeeded")
			}

			if spec.User != nil {
				createUserOptions := usrgrp.CreateUserOptions{
					ID:         spec.User.ID,
					UserGroup:  spec.User.UserGroup,
					Groups:     spec.User.Groups,
					Comment:    spec.User.Comment,
					CreateHome: spec.User.CreateHome,
					Shell:      spec.User.Shell,
				}
				if err = b.CreateUser(spec.User.Name, createUserOptions); err != nil {
					return fmt.Errorf("creating unprivileged user: %w", err)
				}
				logger.Debugf("created unprivileged user '%s'", spec.User.Name)
			}

			if len(spec.Copy) > 0 {
				for _, c := range spec.Copy {
					copyFilesOptions := builder.CopyFilesOptions{
						Excludes: c.Excludes,
						Mode:     c.Mode,
						Owner:    c.Owner,
						RemoveS:  c.RemoveS,
					}
					if err = b.CopyFiles(c.Base, c.Destination, c.Sources, copyFilesOptions); err != nil {
						return fmt.Errorf("copying files to container: %w", err)
					}
				}
				logger.Debugln("file copy step succeeded")
			}

			if spec.Security.SpecialFiles.RemoveS {
				if err = b.UnsetSpecialBits(spec.Security.SpecialFiles.Excludes); err != nil {
					return fmt.Errorf("removing the SUID/SGID bit from files: %w", err)
				}
				logger.Debugln("SUID/SGID bit removal step succeeded")
			}

			if digest != "" {
				spec.Config.Annotations["org.github.ok-ryoko.turret.spec.digest"] = digest
			}

			configureOptions := builder.ConfigureOptions{
				Annotations: spec.Config.Annotations,
				Author:      spec.Config.Author,
				Clear:       spec.From.Clear,
				Command:     spec.Config.Command,
				CreatedBy:   spec.Config.CreatedBy,
				Entrypoint:  spec.Config.Entrypoint,
				Environment: spec.Config.Environment,
				Labels:      spec.Config.Labels,
				Ports:       spec.Config.Ports,
				User:        spec.User,
				WorkDir:     spec.Config.WorkDir,
			}
			b.Configure(configureOptions)
			logger.Debugln("configured image")

			logger.Debugln("committing image...")
			commitOptions := builder.CommitOptions{
				KeepHistory: spec.KeepHistory,
				Latest:      cCtx.Bool("latest"),
			}
			imageID, err := b.Commit(
				ctx,
				store,
				fmt.Sprintf("localhost/%s", spec.Repository),
				spec.Tag,
				commitOptions,
			)
			if err != nil {
				return fmt.Errorf("committing image: %w", err)
			}
			logger.Infof("built and committed %s Linux image %s", distro.String(), imageID)

			return nil
		},
	}
}

// Resolve the absolute path of `p`, assuming `p` refers to a regular file
// and asserting the absolute path is rooted in both the user's home directory
// and the current working directory on the host.
func processPath(p string) (string, error) {
	p, err := filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("determining absolute path: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting user home directory: %w", err)
	}
	if !strings.HasPrefix(p, home) {
		return "", fmt.Errorf("path isn't a child of the home directory")
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting working directory: %w", err)
	}
	if !strings.HasPrefix(p, wd) {
		return "", fmt.Errorf("path isn't a child of the working directory")
	}

	return p, nil
}

// createSpec decodes the contents of the TOML file at the absolute path `p`
// into a build spec, filling in missing values, validating the result, and
// optionally returning an annotated string representation of the file's SHA256
// digest.
func createSpec(p string, hash bool) (spec.Spec, string, error) {
	if !filepath.IsAbs(p) {
		return spec.Spec{}, "", fmt.Errorf("expected absolute path, got %q", p)
	}

	blob, err := os.ReadFile(p)
	if err != nil {
		return spec.Spec{}, "", fmt.Errorf("loading spec: %w", err)
	}

	digest := ""
	if hash {
		h := sha256.New()
		h.Write(blob)
		digest = fmt.Sprintf("sha256:%x", h.Sum(nil))
	}

	r := bytes.NewReader(blob)
	d := toml.NewDecoder(r)
	d.DisallowUnknownFields()

	s := spec.Spec{}
	if err = d.Decode(&s); err != nil {
		return spec.Spec{}, "", fmt.Errorf("decoding TOML: %w", err)
	}

	s.Fill()

	if len(s.Copy) > 0 {
		parent := filepath.Dir(p)
		for i, c := range s.Copy {
			if c.Base == "" {
				s.Copy[i].Base = parent
			} else if filepath.IsLocal(c.Base) {
				s.Copy[i].Base, err = filepath.Abs(filepath.Join(parent, c.Base))
				if err != nil {
					return spec.Spec{}, "", fmt.Errorf("canonicalizing base path %q", c.Base)
				}
			}
		}
	}

	if err = s.Validate(); err != nil {
		return spec.Spec{}, "", fmt.Errorf("validating spec: %w", err)
	}

	return s, digest, nil
}

func setLoggerLevel(l *logrus.Logger, verbosity uint) {
	switch verbosity {
	case 0:
		l.SetLevel(logrus.ErrorLevel)
	case 1:
		l.SetLevel(logrus.WarnLevel)
	case 2:
		l.SetLevel(logrus.InfoLevel)
	default:
		l.SetLevel(logrus.DebugLevel)
	}
}
