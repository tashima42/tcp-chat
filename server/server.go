package server

import (
	"bufio"
	"log"
	"net"
	"sync"

	"github.com/google/uuid"
	"github.com/tashima42/tcp-chat/types"
	"github.com/urfave/cli/v2"
)

const (
	protocol = "tcp"
)

type user struct {
	username string
	conn     net.Conn
}

func Command() *cli.Command {
	return &cli.Command{
		Name:  "server",
		Usage: "tcp chat server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "address",
				Usage:    "address to bind the server to",
				Aliases:  []string{"a"},
				Required: true,
			},
		},
		Action: serverCommand,
	}
}

func serverCommand(ctx *cli.Context) error {
	address := ctx.String("address")
	return server(address)
}

func server(address string) error {
	listen, err := net.Listen(protocol, address)
	if err != nil {
		return err
	}
	defer listen.Close()

	var connMap = &sync.Map{}
	for {
		conn, err := listen.Accept()
		if err != nil {
			return err
		}

		id := uuid.New().String()
		u := user{conn: conn, username: ""}
		connMap.Store(id, conn)

		go handleConnection(id, u, connMap)
	}
}

func handleConnection(id string, u user, connMap *sync.Map) {
	defer func() {
		u.conn.Close()
		connMap.Delete(id)
	}()

	for {
		input, err := bufio.NewReader(u.conn).ReadBytes('\n')
		if err != nil {
			log.Print("Error reading action: " + err.Error())
			return
		}

		action := types.Action{}
		_, err = action.UnmarshalMsg(input)
		if err != nil {
			log.Print("Error unmarshalling action: " + err.Error())
			return
		}

		actionType := types.ActionType(action.Type)
		if actionType != types.ActionTypeRegister && u.username == "" {
			errMsg := types.ErrorMessage{Value: "user must be registered before sending messages"}
			var errB []byte
			if _, err := errMsg.MarshalMsg(errB); err != nil {
				log.Print("Failed to marshall error message: " + err.Error())
			}
			if _, err := u.conn.Write(errB); err != nil {

			}
		}

		switch actionType {
		case types.ActionTypeRegister:

		case types.ActionTypeMessage:
		}

		connMap.Range(func(key, value interface{}) bool {
			if key == id {
				return true
			}
			if conn, ok := value.(net.Conn); ok {
				log.Printf("[%s]: %s", id, input)
				if _, err := conn.Write([]byte(input)); err != nil {
					log.Print("Error writing to connection " + err.Error())
				}
			}
			return true
		})
	}
}
