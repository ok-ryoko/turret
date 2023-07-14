// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/containers/buildah"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

func main() {
	if buildah.InitReexec() {
		return
	}
	logger := newLogger()
	app := newApp(logger)
	customizeHelpTemplates()
	if err := app.Run(os.Args); err != nil {
		logger.Fatalf("%s crashed: %v", app.HelpName, err)
	}
}

func newApp(logger *logrus.Logger) *cli.App {
	return &cli.App{
		Name:           "Turret",
		HelpName:       "turret",
		Usage:          "Build rootless OCI images of Linux-based distros declaratively",
		Version:        "0.1.0",
		DefaultCommand: "help",
		Commands: []*cli.Command{
			newBuildCmd(logger),
			newVersionCmd(),
		},
		HideVersion: true,
		Authors: []*cli.Author{
			{
				Name:  "OK Ryoko",
				Email: "ryoko@kyomu.jp.net",
			},
		},
		Copyright: "(c) 2023 OK Ryoko",
	}
}
