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
			localRef := filepath.Join("localhost", imageRef)

			if store.Exists(localRef) && !cCtx.Bool("force") {
				return fmt.Errorf("image %s already exists", localRef)
			}

			distro := spec.Distro.Distro
			distroString := distro.String()
			packageManager := spec.Packages.Manager.PackageManager
			baseRef := spec.From.Reference()
			commonOptions := builder.CommonOptions{
				LogCommands: v >= 4,
			}
			b, err := builder.New(
				ctx,
				distro,
				packageManager,
				baseRef,
				cCtx.Bool("pull"),
				store,
				logger,
				commonOptions,
			)
			if err != nil {
				return fmt.Errorf("creating %s Linux Turret builder: %w", distroString, err)
			}
			defer func() {
				if !cCtx.Bool("keep") {
					if removeErr := b.Remove(); removeErr != nil {
						logger.Warnln("failed deleting working container")
						logger.Infoln("please remove the container manually: buildah rm", b.ContainerID())
					}
				}
			}()
			logger.Debugf("created %s Linux Turret builder", distroString)

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
				createUserOptions := builder.CreateUserOptions{
					ID:         spec.User.ID,
					UserGroup:  spec.User.UserGroup,
					Groups:     spec.User.Groups,
					Comment:    spec.User.Comment,
					LoginShell: spec.User.LoginShell,
				}
				if err = b.CreateUser(spec.User.Name, spec.Distro.Distro, createUserOptions); err != nil {
					return fmt.Errorf("creating unprivileged user: %w", err)
				}
				logger.Debugf("created unprivileged user '%s'", spec.User.Name)
			}

			if len(spec.Copy) > 0 {
				copyFilesOptions := builder.CopyFilesOptions{
					UserName: spec.User.Name,
				}
				if err = b.CopyFiles(spec.Copy, copyFilesOptions); err != nil {
					return fmt.Errorf("copying some files in home directory: %w", err)
				}
				logger.Debugln("file copy step succeeded")
			}

			if spec.Security.SpecialFiles.RemoveS {
				if err = b.UnsetSpecialBits(spec.Security.SpecialFiles.Excludes); err != nil {
					return fmt.Errorf("removing SUID and SGID bits on binaries: %w", err)
				}
				logger.Debugln("SUID/SGID bit removal step succeeded")
			}

			if len(digest) > 0 {
				spec.Annotations["org.github.ok-ryoko.turret.spec.digest"] = digest
			}

			configureOptions := builder.ConfigureOptions{
				Annotations: spec.Annotations,
				Env:         spec.Env,
				LoginShell:  spec.User.LoginShell,
				UserName:    spec.User.Name,
			}
			b.Configure(spec.User != nil, configureOptions)
			logger.Debugln("configured image")

			logger.Debugln("committing image...")
			commitOptions := builder.CommitOptions{
				KeepHistory: spec.KeepHistory,
				Latest:      cCtx.Bool("latest"),
			}
			imageId, err := b.Commit(
				ctx,
				store,
				fmt.Sprintf("localhost/%s", spec.Repository),
				spec.Tag,
				commitOptions,
			)
			if err != nil {
				return fmt.Errorf("committing image: %w", err)
			}
			logger.Infof("built and committed %s Linux image %s", distroString, imageId)

			return nil
		},
	}
}

// Resolve the absolute path of `p`;
// assume the path refers to a regular file;
// assert the path is rooted in both the user's home directory and the current
// working directory
func processPath(p string) (string, error) {
	p, err := filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("determining absolute path: %w", err)
	}

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting user home directory: %w", err)
	}
	if !strings.HasPrefix(p, userHomeDir) {
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

// Decode the contents of the TOML file at `p` into a Turret build spec;
// normalize and validate the spec;
// optionally, compute the SHA256 digest of the spec file
func createSpec(p string, hash bool) (builder.Spec, string, error) {
	blob, err := os.ReadFile(p)
	if err != nil {
		return builder.Spec{}, "", fmt.Errorf("loading spec: %w", err)
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

	spec := builder.Spec{}
	if err = d.Decode(&spec); err != nil {
		return builder.Spec{}, "", fmt.Errorf("decoding TOML: %w", err)
	}

	spec.Fill()

	if err = spec.Validate(); err != nil {
		return builder.Spec{}, "", fmt.Errorf("validating spec: %w", err)
	}

	return spec, digest, nil
}

func setLoggerLevel(logger *logrus.Logger, verbosity uint) {
	switch verbosity {
	case 0:
		logger.SetLevel(logrus.ErrorLevel)
	case 1:
		logger.SetLevel(logrus.WarnLevel)
	case 2:
		logger.SetLevel(logrus.InfoLevel)
	default:
		logger.SetLevel(logrus.DebugLevel)
	}
}
