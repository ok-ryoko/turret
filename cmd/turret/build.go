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

	"github.com/ok-ryoko/turret/pkg/build"
	"github.com/ok-ryoko/turret/pkg/spec"

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

			verbosity := cCtx.Uint("verbosity")
			if cCtx.Bool("quiet") {
				verbosity = 0
			}
			setLoggerLevel(logger, verbosity)

			specPath, err := filepath.Abs(cCtx.Args().First())
			if err != nil {
				return fmt.Errorf("canonicalizing spec path: %w", err)
			}
			logger.Debugln("processed spec path")

			spec, digest, err := createSpec(specPath, cCtx.Bool("hash-spec"))
			if err != nil {
				return fmt.Errorf("creating in-memory representation of spec: %w", err)
			}
			logger.Debugln("created in-memory representation of spec")

			options := build.ExecuteOptions{
				Digest:      digest,
				Force:       cCtx.Bool("force"),
				Keep:        cCtx.Bool("keep"),
				Latest:      cCtx.Bool("latest"),
				LogCommands: verbosity >= 4,
				Pull:        cCtx.Bool("pull"),
			}
			if err := build.Execute(ctx, spec, logger, options); err != nil {
				return fmt.Errorf("building image according to given spec: %w", err)
			}

			return nil
		},
	}
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
