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

	users := map[string]types.User{}

	var connMap = &sync.Map{}
	for {
		conn, err := listen.Accept()
		if err != nil {
			return err
		}

		u := types.NewUser(uuid.New().String(), "", conn)
		connMap.Store(u.ID, conn)

		go handleConnection(u, connMap, &users)
	}
}

func handleConnection(u types.User, connMap *sync.Map, users *map[string]types.User) {
	defer func() {
		u.GetConn().Close()
		delete((*users), u.ID)
		connMap.Delete(u.ID)
	}()

	for {
		input, err := bufio.NewReader(u.GetConn()).ReadBytes('\n')
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
		if actionType != types.ActionTypeRegister && u.Username == "" {
			errMsg := types.ErrorMessage{Value: "user must be registered before sending messages"}
			var errB []byte
			if _, err := errMsg.MarshalMsg(errB); err != nil {
				log.Print("Failed to marshall error message: " + err.Error())
				return
			}
			errB = append(errB, '\n')
			if _, err := u.GetConn().Write(errB); err != nil {
				log.Print("Failed to write error message: " + err.Error())
				return
			}
		}

		switch actionType {
		case types.ActionTypeRegister:
			user := types.User{}
			if _, err = user.UnmarshalMsg(action.Data); err != nil {
				log.Print("Error unmarshalling register: " + err.Error())
				return
			}
			log.Println("Registering user: " + user.Username)
			u.Username = user.Username
			user.ID = u.ID
			(*users)[u.ID] = user
			sendUsers := types.Users{}
			for _, v := range *users {
				sendUsers = append(sendUsers, v)
			}
			usersB, _ := sendUsers.MarshalMsg(nil)
			sendActions(u.ID, connMap, types.ActionTypeGetUsers, usersB)
		case types.ActionTypeMessage:
			message := types.Message{}
			if _, err = message.UnmarshalMsg(action.Data); err != nil {
				log.Print("Error unmarshalling message: " + err.Error())
				return
			}
			message.UserID = u.ID
			messageB, _ := message.MarshalMsg(nil)
			log.Printf("Recieved message: %+v", message)
			sendActions(u.ID, connMap, types.ActionTypeMessage, messageB)
		}
	}
}

func sendActions(id string, connMap *sync.Map, actionType types.ActionType, data []byte) {
	action := types.Action{
		Type: actionType,
		Data: data,
	}
	actionB, _ := action.MarshalMsg(nil)
	actionB = append(actionB, '\n')
	connMap.Range(func(key, value interface{}) bool {
		if key == id && actionType != types.ActionTypeGetUsers {
			return true
		}
		if conn, ok := value.(net.Conn); ok {
			if _, err := conn.Write(actionB); err != nil {
				log.Print("Error writing to connection " + err.Error())
			}
		}
		return true
	})
}
