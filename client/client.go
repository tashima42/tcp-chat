package client

import (
	"bufio"
	"fmt"
	"net"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tashima42/tcp-chat/types"
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
	p := tea.NewProgram(initialModel(conn), tea.WithAltScreen())
	go read(*conn, p)
	if _, err := p.Run(); err != nil {
		fmt.Println("Uh oh", err)
		os.Exit(1)
	}

	return nil
}

func wrapAction(actionType types.ActionType, data []byte) []byte {
	action := types.Action{
		Type: actionType,
		Data: data,
	}
	actionB, _ := action.MarshalMsg(nil)
	return actionB
}

func register(conn net.Conn, username string) {
	registerMsg := types.User{Username: username}
	registerB, _ := registerMsg.MarshalMsg(nil)
	actionB := wrapAction(types.ActionTypeRegister, registerB)
	write(conn, actionB)
}
func sendMessage(conn net.Conn, value string) {
	msg := types.Message{Value: value}
	msgB, _ := msg.MarshalMsg(nil)
	actionB := wrapAction(types.ActionTypeMessage, msgB)
	write(conn, actionB)
}

func connect(address string) (*net.Conn, error) {
	conn, err := net.Dial(protocol, address)
	return &conn, err
}

func write(conn net.Conn, content []byte) error {
	content = append(content, '\n')
	if _, err := conn.Write(content); err != nil {
		return err
	}
	return nil
}

func read(conn net.Conn, p *tea.Program) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		action := types.Action{}
		action.UnmarshalMsg(scanner.Bytes())
		switch types.ActionType(action.Type) {
		case types.ActionTypeMessage:
			msg := types.Message{}
			msg.UnmarshalMsg(action.Data)
			p.Send(msg)
		case types.ActionTypeGetUsers:
			users := types.Users{}
			users.UnmarshalMsg(action.Data)
			p.Send(users)
		}
	}
}
