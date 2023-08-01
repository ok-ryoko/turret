// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/containers/buildah"
	"github.com/sirupsen/logrus"
)

// Container represents a working container.
type Container struct {
	// Pointer to the underlying Buildah builder instance
	Builder *buildah.Builder

	// Pointer to the underlying logger
	Logger *logrus.Logger

	// Common options for the execution of any container process
	CommonOptions CommonOptions
}

// CommonOptions holds options for the execution of any container process.
type CommonOptions struct {
	// Environment variables to set when running a command in the working
	// container, represented as a slice of "KEY=VALUE"s
	Env []string

	// Whether to log the output and error streams of container processes
	LogCommands bool
}

// ContainerID returns the ID of the working container.
func (c *Container) ContainerID() string {
	return buildah.GetBuildInfo(c.Builder).ContainerID
}

// defaultRunOptions instantiates a buildah.RunOptions from the container's
// common execution options.
func (c *Container) DefaultRunOptions() buildah.RunOptions {
	ro := buildah.RunOptions{
		ConfigureNetwork: buildah.NetworkDisabled,
		Quiet:            true,
	}

	if len(c.CommonOptions.Env) > 0 {
		ro.Env = append(ro.Env, c.CommonOptions.Env...)
	}

	if c.CommonOptions.LogCommands {
		ro.Logger = c.Logger
		ro.Quiet = false
	}

	return ro
}

// Remove removes the working container and destroys this Container, which
// should not be used afterwards.
func (c *Container) Remove() error {
	err := c.Builder.Delete()
	if err != nil {
		return fmt.Errorf("deleting container: %w", err)
	}
	*c = Container{}
	return nil
}

// resolveExecutable returns the absolute path of an executable in the working
// container if it can be found and an error otherwise, assuming `command` can
// be resolved.
func (c *Container) ResolveExecutable(executable string) (string, error) {
	cmd := []string{"/bin/sh", "-c", "command", "-v", executable}

	var buf bytes.Buffer
	ro := c.DefaultRunOptions()
	ro.Stdout = &buf

	if err := c.Run(cmd, ro); err != nil {
		return "", fmt.Errorf("running default shell or resolving executable '%s'", executable)
	}

	return strings.TrimSpace(buf.String()), nil
}

// run runs a command in the working container, optionally sanitizing and
// logging the process's standard output and error streams. When sanitizing, it
// strips all ANSI escape codes as well as superfluous whitespace.
func (c *Container) Run(cmd []string, options buildah.RunOptions) error {
	var stderrBuf bytes.Buffer
	if options.Stderr == nil && c.CommonOptions.LogCommands {
		options.Stderr = &stderrBuf
	}

	var stdoutBuf bytes.Buffer
	if options.Stdout == nil && c.CommonOptions.LogCommands {
		options.Stdout = &stdoutBuf
	}

	defer func() {
		if c.CommonOptions.LogCommands {
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
					c.Logger.Debugf("%s: stderr: %s", cmd[0], l)
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
					c.Logger.Debugf("%s: stdout: %s", cmd[0], l)
				}
			}
		}
	}()

	if err := c.Builder.Run(cmd, options); err != nil {
		return fmt.Errorf("%w", err)
	}

	return nil
}
