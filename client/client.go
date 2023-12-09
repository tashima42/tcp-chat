package client

import (
	"bufio"
	"fmt"
	"net"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/urfave/cli/v2"
)

const (
	protocol = "tcp"
)

func Command() *cli.Command {
	return &cli.Command{
		Name:  "client",
		Usage: "tcp chat client",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "address",
				Usage:    "address to bind the server to",
				Aliases:  []string{"a"},
				Required: true,
			},
		},
		Action: clientCommand,
	}
}

func clientCommand(ctx *cli.Context) error {
	address := ctx.String("address")
	conn, err := connect(address)
	if err != nil {
		return err
	}
	p := tea.NewProgram(initialModel(conn))
	go read(*conn, p)
	if _, err := p.Run(); err != nil {
		fmt.Println("Uh oh", err)
		os.Exit(1)
	}

	return nil
}

func connect(address string) (*net.Conn, error) {
	conn, err := net.Dial(protocol, address)
	return &conn, err
}

func write(conn net.Conn, message string) error {
	if _, err := conn.Write([]byte(message + "\n")); err != nil {
		return err
	}
	return nil
}

func read(conn net.Conn, p *tea.Program) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		p.Send(newMsg{value: scanner.Text()})
	}
}
