package main

import (
	"os"

	"github.com/upccup/july/command"

	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "july"
	app.Version = "1.0.0"
	app.Author = "upccup"
	app.Usage = "docker network plugin with remote IPAM & event listener"
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "cluster-store", Value: "http://127.0.0.1:2379", Usage: "the key/value store endpoint url. [$CLUSTER_STORE]"},
		cli.BoolFlag{Name: "debug", Usage: "debug mode [$DEBUG]"},
	}
	app.Commands = []cli.Command{
		command.NewServerCommand(),
		command.NewIPRangeCommand(),
		command.NewReleaseIPCommand(),
		command.NewHostRangeCommand(),
		command.NewReleaseHostCommand(),
		command.NewCreateNetworkCommand(),
		command.NewDockerAgentCommand(),
	}
	app.Run(os.Args)
}
