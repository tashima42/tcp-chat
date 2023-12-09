package main

import (
	"log"
	"os"

	"github.com/tashima42/tcp-chat/client"
	"github.com/tashima42/tcp-chat/server"
	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.App{
		Name:  "tcp-chat",
		Usage: "tcp chat server and TUI client",
		Commands: []*cli.Command{
			server.Command(),
			client.Command(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
