package types

import "net"

type ActionType int

const (
	ActionTypeRegister ActionType = 1
	ActionTypeMessage  ActionType = 2
	ActionTypeGetUsers ActionType = 3
)

//go:generate msgp
type Action struct {
	Type ActionType //`msg:"type"`
	Data []byte     //`msg:"data"`
}

type User struct {
	ID       string   //`msg:"id"`
	Username string   //`msg:"username"`
	conn     net.Conn //`msg:"-"`
}
type Users []User

func NewUser(id, username string, conn net.Conn) User {
	return User{
		ID:       id,
		Username: username,
		conn:     conn,
	}
}

func (u *User) GetConn() net.Conn {
	return u.conn
}

type Message struct {
	UserID string //`msg:"userId"`
	Value  string //`msg:"value"`
}

type ErrorMessage struct {
	Value string //`msg:"value"`
}
