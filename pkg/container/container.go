// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package container

import (
	"bytes"
	"fmt"
	"strings"

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
	if err := c.Builder.Delete(); err != nil {
		return fmt.Errorf("deleting container %s: %w", c.ContainerID(), err)
	}
	*c = Container{}
	return nil
}

// resolveExecutable returns the absolute path of an executable in the working
// container if it can be found and an error otherwise, assuming `command` can
// be resolved.
func (c *Container) ResolveExecutable(executable string) (string, error) {
	cmd := []string{"command", "-v", executable}
	ro := c.DefaultRunOptions()
	resolved, _, err := c.Run(cmd, ro)
	if err != nil {
		return "", fmt.Errorf("resolving executable %q: %w", executable, err)
	}
	return strings.TrimSpace(resolved), nil
}

// Run executes a command in the working container, capturing standard output
// and standard error streams as UTF-8-encoded strings.
func (c *Container) Run(cmd []string, options buildah.RunOptions) (string, string, error) {
	var (
		stdoutBuf bytes.Buffer
		stderrBuf bytes.Buffer
	)

	options.Stdout = &stdoutBuf
	options.Stderr = &stderrBuf
	err := c.Builder.Run(cmd, options)

	outText := stdoutBuf.String()
	if outText == "<nil>" {
		outText = ""
	}

	errText := stderrBuf.String()
	if errText == "<nil>" {
		errText = ""
	}

	if err != nil {
		return outText, errText, fmt.Errorf("%w", err)
	}
	return outText, errText, nil
}

// runWithLogging wraps Run, logging standard output and standard error.
func (c *Container) runWithLogging(cmd []string, options buildah.RunOptions, errContext string) error {
	outText, errText, err := c.Run(cmd, options)
	if err != nil {
		if errText != "" {
			errContext = fmt.Sprintf("%s (%q)", errContext, errText)
		}
		return fmt.Errorf("%s: %w", errContext, err)
	}
	if errText != "" {
		c.Logger.Warn(errText)
	}
	if c.CommonOptions.LogCommands {
		if outText != "" {
			c.Logger.Debug(outText)
		}
	}
	return nil
}
