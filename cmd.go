package main

import (
	"os"

	"github.com/teris-io/cli"
)

func parse() {
	co := cli.NewCommand("checkout", "checkout a branch or revision").
		WithShortcut("co").
		WithArg(cli.NewArg("revision", "branch or revision to checkout")).
		WithOption(cli.NewOption("branch", "Create branch if missing").WithChar('b').WithType(cli.TypeBool)).
		WithOption(cli.NewOption("upstream", "Set upstream for the branch").WithChar('u').WithType(cli.TypeBool)).
		WithAction(func(args []string, options map[string]string) int {
			// do something
			return 0
		})

	add := cli.NewCommand("add", "add a remote").WithArg(cli.NewArg("remote", "remote to add"))

	rmt := cli.NewCommand("remote", "Work with git remotes").WithCommand(add)

	app := cli.New("git tool").
		WithOption(cli.NewOption("verbose", "Verbose execution").WithChar('v').WithType(cli.TypeBool)).
		WithCommand(co).
		WithCommand(rmt)
	// no action attached, just print usage when executed

	os.Exit(app.Run(os.Args, os.Stdout))
}
