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
		if _, err = action.UnmarshalMsg(input); err != nil {
			log.Print("Error unmarshalling action: " + err.Error())
			return
		}

		actionType := types.ActionType(action.Type)
		if actionType != types.ActionTypeRegister && u.username == "" {
			errMsg := types.ErrorMessage{Value: "user must be registered before sending messages"}
			var errB []byte
			if _, err := errMsg.MarshalMsg(errB); err != nil {
				log.Print("Failed to marshall error message: " + err.Error())
				return
			}
			errB = append(errB, '\n')
			if _, err := u.conn.Write(errB); err != nil {
				log.Print("Failed to write error message: " + err.Error())
				return
			}
		}

		switch actionType {
		case types.ActionTypeRegister:
			register := types.Register{}
			if _, err = register.UnmarshalMsg(action.Data); err != nil {
				log.Print("Error unmarshalling register: " + err.Error())
				return
			}
			log.Println("Registering user: " + register.Username)
			u.username = register.Username
		case types.ActionTypeMessage:
			message := types.Message{}
			if _, err = message.UnmarshalMsg(action.Data); err != nil {
				log.Print("Error unmarshalling message: " + err.Error())
				return
			}
			message.ID = id
			message.Username = u.username
			messageB, _ := message.MarshalMsg(nil)
			messageB = append(messageB, '\n')
			log.Printf("Recieved message: %+v", message)
			connMap.Range(func(key, value interface{}) bool {
				if key == id {
					return true
				}
				if conn, ok := value.(net.Conn); ok {
					if _, err := conn.Write(messageB); err != nil {
						log.Print("Error writing to connection " + err.Error())
					}
				}
				return true
			})
		}
	}
}
