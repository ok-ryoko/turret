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
	"github.com/ok-ryoko/turret/pkg/linux/user"
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
				return fmt.Errorf("processing spec path: %w", err)
			}
			logger.Debugln("processed spec path")

			spec, digest, err := createSpec(specPath, cCtx.Bool("hash-spec"))
			if err != nil {
				return fmt.Errorf("creating in-memory representation of spec: %w", err)
			}
			logger.Debugln("created in-memory representation of spec")

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

			refThis := spec.This.Reference()
			if store.Exists(refThis) && !cCtx.Bool("force") {
				return fmt.Errorf("image %s already exists", refThis)
			}

			distro := spec.This.Distro.Distro
			b, err := builder.New(
				ctx,
				store,
				spec.From.Reference(),
				logger,
				cCtx.Bool("pull"),
				distro,
				spec.Backends.Package.Manager,
				spec.Backends.User.Manager,
				spec.Backends.Finder.Finder,
				container.CommonOptions{LogCommands: v >= 4},
			)
			if err != nil {
				return fmt.Errorf("creating %s Linux working container: %w", distro, err)
			}
			defer func() {
				if !cCtx.Bool("keep") {
					if removeErr := b.Remove(); removeErr != nil {
						logger.Warnln("failed deleting working container")
						logger.Infoln("please remove the container manually: buildah rm", b.ContainerID())
					}
				}
			}()
			logger.Debugf("created %s Linux working container", distro)

			if b.Builder.OS() != "linux" {
				return fmt.Errorf("expected 'linux' image, got '%s' image", b.Builder.OS())
			}

			if spec.Packages.Upgrade {
				logger.Debugln("upgrading packages in the working container...")
				if err = b.UpgradePackages(); err != nil {
					return fmt.Errorf("upgrading packages: %w", err)
				}
				logger.Debugln("upgrade command ran successfully")
			}

			if len(spec.Packages.Install) > 0 {
				logger.Debugln("installing packages to the working container...")
				if err = b.InstallPackages(spec.Packages.Install); err != nil {
					return fmt.Errorf("installing packages: %w", err)
				}
				logger.Debugln("install command ran successfully")
			}

			if spec.Packages.Clean {
				if err = b.CleanPackageCaches(); err != nil {
					return fmt.Errorf("cleaning package caches: %w", err)
				}
				logger.Debugln("clean command ran successfully")
			}

			if spec.User != nil {
				createUserOptions := user.CreateUserOptions{
					ID:         spec.User.ID,
					UserGroup:  spec.User.UserGroup,
					Groups:     spec.User.Groups,
					Comment:    spec.User.Comment,
					CreateHome: spec.User.CreateHome,
				}
				if err = b.CreateUser(spec.User.Name, createUserOptions); err != nil {
					return fmt.Errorf("creating nonroot user: %w", err)
				}
				logger.Debugf("created nonroot user")
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
						return fmt.Errorf("copying files: %w", err)
					}
				}
				logger.Debugln("file copy command(s) ran successfully")
			}

			if spec.Security.SpecialFiles.RemoveS {
				if err = b.UnsetSpecialBits(spec.Security.SpecialFiles.Excludes); err != nil {
					return fmt.Errorf("removing SUID and SGID bits from files: %w", err)
				}
				logger.Debugln("command to remove SUID and SGID bits from files ran successfully")
			}

			if digest != "" {
				spec.Config.Annotations["org.github.ok-ryoko.turret.spec.digest"] = digest
			}

			ports := make([]string, len(spec.Config.Ports))
			for i, p := range spec.Config.Ports {
				ports[i] = p.String()
			}

			var configureUserOptions builder.ConfigureUserOptions
			if spec.User != nil {
				configureUserOptions = builder.ConfigureUserOptions{
					Name:       spec.User.Name,
					CreateHome: spec.User.CreateHome,
				}
			}

			configureOptions := builder.ConfigureOptions{
				ClearAnnotations: spec.Config.Clear.Annotations,
				Annotations:      spec.Config.Annotations,
				ClearAuthor:      spec.Config.Clear.Author,
				Author:           spec.Config.Author,
				ClearCommand:     spec.Config.Clear.Command,
				Command:          spec.Config.Command,
				CreatedBy:        spec.Config.CreatedBy,
				ClearEntrypoint:  spec.Config.Clear.Entrypoint,
				Entrypoint:       spec.Config.Entrypoint,
				ClearEnvironment: spec.Config.Clear.Environment,
				Environment:      spec.Config.Environment,
				ClearLabels:      spec.Config.Clear.Labels,
				Labels:           spec.Config.Labels,
				ClearPorts:       spec.Config.Clear.Ports,
				Ports:            ports,
				User:             &configureUserOptions,
				WorkDir:          spec.Config.WorkDir,
			}
			b.Configure(configureOptions)
			logger.Debugln("configured image")

			logger.Debugln("committing image...")
			commitOptions := builder.CommitOptions{
				KeepHistory: spec.This.KeepHistory,
				Latest:      cCtx.Bool("latest"),
			}
			imageID, err := b.Commit(
				ctx,
				store,
				spec.This.Repository,
				spec.This.Tag,
				commitOptions,
			)
			if err != nil {
				return fmt.Errorf("committing image: %w", err)
			}
			logger.Infoln(imageID)

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
		return "", fmt.Errorf("canonicalizing spec path: %w", err)
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
		return spec.Spec{}, "", fmt.Errorf("reading spec file: %w", err)
	}

	digest := ""
	if hash {
		digest = fmt.Sprintf("sha256:%x", sha256.Sum256(blob))
	}

	r := bytes.NewReader(blob)
	d := toml.NewDecoder(r)
	d.DisallowUnknownFields()

	s := spec.Spec{}
	if err = d.Decode(&s); err != nil {
		return spec.Spec{}, "", fmt.Errorf("decoding TOML: %w", err)
	}

	s = spec.Fill(s)

	if len(s.Copy) > 0 {
		parent := filepath.Dir(p)
		for i, c := range s.Copy {
			if c.Base == "" {
				s.Copy[i].Base = parent
			} else if strings.HasPrefix(c.Base, "~") {
				home, err := os.UserHomeDir()
				if err != nil {
					return spec.Spec{}, "", fmt.Errorf("discovering home directory on host: %w", err)
				}
				if c.Base == "~" {
					s.Copy[i].Base = home
				} else if strings.HasPrefix(c.Base, "~/") {
					_, after, _ := strings.Cut(c.Base, "/")
					s.Copy[i].Base = filepath.Clean(filepath.Join(home, after))
				}
			} else if filepath.IsLocal(c.Base) {
				s.Copy[i].Base, err = filepath.Abs(filepath.Join(parent, c.Base))
				if err != nil {
					return spec.Spec{}, "", fmt.Errorf("canonicalizing base path %q", c.Base)
				}
			} else {
				s.Copy[i].Base = filepath.Clean(c.Base)
			}

			if c.Destination != "" {
				s.Copy[i].Destination = filepath.Clean(c.Destination)
			}

			for j, src := range s.Copy[i].Sources {
				if src != "" {
					s.Copy[i].Sources[j] = filepath.Clean(src)
				}
			}
		}
	}

	if err = spec.Validate(s); err != nil {
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
