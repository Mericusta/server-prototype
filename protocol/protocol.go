package protocol

import (
	"server-prototype/model"
)

// TCPPacketDataTagSize TCP 包中数据的标识的值的占位长度
const TCPPacketDataTagSize = 4

// TCPPacketDataLengthSize TCP 包中数据的长度的值的占位长度
const TCPPacketDataLengthSize = 4

type ServerState int

const (
	Shutdown    ServerState = 1
	Running     ServerState = 2
	Maintaining ServerState = 3
)

type MsgID int

type TLVMessage struct {
	Tag    MsgID
	Length int
	Value  []byte
}

type IMessage interface {
	Tag() MsgID
}

const MessageEcho MsgID = 0

type C2SEcho struct {
	Value int
}

func (message C2SEcho) Tag() MsgID {
	return MessageEcho
}

type S2CEcho struct {
	Value int
}

func (message S2CEcho) Tag() MsgID {
	return MessageEcho
}

const MessageOnline MsgID = 1

type C2SOnline struct {
}

func (message C2SOnline) Tag() MsgID {
	return MessageOnline
}

type S2COnline struct {
	State ServerState
	Error string
}

func (message S2COnline) Tag() MsgID {
	return MessageOnline
}

const MessageOffline MsgID = 2

type C2SOffline struct {
}

func (message C2SOffline) Tag() MsgID {
	return MessageOffline
}

type S2COffline struct {
	ServerState int
	Error       string
}

func (message S2COffline) Tag() MsgID {
	return MessageOffline
}

const MessageLogin MsgID = 3

type C2SLogin struct {
	Username string
	Password string
}

func (message C2SLogin) Tag() MsgID {
	return MessageLogin
}

type S2CLogin struct {
	Player         model.PlayerData
	NeedCreateRole bool
}

func (message S2CLogin) Tag() MsgID {
	return MessageLogin
}
