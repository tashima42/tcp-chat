package types

type ActionType int

const (
	ActionTypeRegister ActionType = 1
	ActionTypeMessage  ActionType = 2
)

//go:generate msgp
type Action struct {
	Type ActionType //`msg:"type"`
	Data []byte     //`msg:"data"`
}

type Register struct {
	Username string //`msg:"username"`
}

type RegisterResponse struct {
	ID string //`msg:"id"`
}

type Message struct {
	ID       string //`msg:"id"`
	Username string //`msg:"username"`
	Value    string //`msg:"value"`
}

type ErrorMessage struct {
	Value string //`msg:"value"`
}
