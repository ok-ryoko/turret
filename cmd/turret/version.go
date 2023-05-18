package main

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func newVersionCmd() *cli.Command {
	return &cli.Command{
		Name:            "version",
		Aliases:         []string{"v"},
		Usage:           "Print program version number and exit",
		HideHelpCommand: true,
		Action: func(cCtx *cli.Context) error {
			fmt.Println(cCtx.App.Version)
			return nil
		},
	}
}
