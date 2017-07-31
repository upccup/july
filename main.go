package main

import (
	"os"

	"github.com/upccup/july/command"
	"github.com/upccup/july/db"

	log "github.com/Sirupsen/logrus"
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
	app.Before = InitConfig
	app.Commands = []cli.Command{
		command.NewServerCommand(),
		command.NewIPRangeCommand(),
		command.NewReleaseIPCommand(),
		command.NewHostRangeCommand(),
		command.NewReleaseHostCommand(),
		command.NewCreateNetworkCommand(),
	}
	app.Run(os.Args)
}

func InitConfig(c *cli.Context) error {
	initialize_log(c.GlobalBool("debug"))

	log.Info("cluster-store endpoint: ", c.GlobalString("cluster-store"))
	db.SetDBAddr(c.GlobalString("cluster-store"))
	return nil
}

func initialize_log(debug bool) {
	log.SetOutput(os.Stderr)
	if debug {
		log.Info("set log level to debug")
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}
