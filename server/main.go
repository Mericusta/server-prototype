package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"sync/atomic"
	"time"

	"server-prototype/model"
	"server-prototype/net"
	"server-prototype/protocol"
)

// global value

var address string = "127.0.0.1:6666"
var clientCount int32

// var clientConnectorMap sync.Map

var serverState atomic.Value
var msgCallbackMap map[protocol.MsgID]func(*net.Connector, []byte)

// test data

var pprof bool = true
var pprofAddress string = "127.0.0.1:8888"
var delayRunningSeconds int = 0

// server process

func init() {
	serverState.Store(protocol.Shutdown)
	msgCallbackMap = make(map[protocol.MsgID]func(*net.Connector, []byte))
	registCallback()
	serverState.Store(protocol.Maintaining)
}

func registCallback() {
	msgCallbackMap[protocol.MessageEcho] = c2sEcho
	msgCallbackMap[protocol.MessageLogin] = c2sLogin
}

func startServer() {
	time.Sleep(time.Second * time.Duration(delayRunningSeconds))
	serverState.Store(protocol.Running)
}

func readPorcess(connector *net.Connector, tag protocol.MsgID, byteDate []byte) {
	var clientNo int32
	if tag == protocol.MessageOnline {
		clientOnline(connector, byteDate)
	} else if tag == protocol.MessageOffline {
		clientOffline(clientNo, connector, byteDate)
	} else {
		messageCallback, hasCallback := msgCallbackMap[tag]
		if !hasCallback {
			fmt.Printf("Warn: message %v does not have callback\n", tag)
			return
		}
		messageCallback(connector, byteDate)
	}
}

// build connection

func clientOnline(connector *net.Connector, msgByte []byte) {
	c2sConnect := &protocol.C2SOnline{}
	unmarshalError := json.Unmarshal(msgByte, c2sConnect)
	if unmarshalError != nil {
		fmt.Printf("Error: unmarshal c2sOnline message error: %v", unmarshalError)
		return
	}
	count := atomic.AddInt32(&clientCount, 1)
	fmt.Printf("Note: A new client is online, client count = %v\n", count)
}

func clientOffline(clientNo int32, connector *net.Connector, msgByte []byte) {
	c2sConnect := &protocol.C2SOffline{}
	unmarshalError := json.Unmarshal(msgByte, c2sConnect)
	if unmarshalError != nil {
		fmt.Printf("Error: unmarshal c2sOffline message error: %v", unmarshalError)
		return
	}
	connector.Close()
	count := atomic.AddInt32(&clientCount, -1)
	fmt.Printf("Note: A client is offline, client count = %v\n", count)
}

// operation

// request callback

func c2sLogin(connector *net.Connector, msgByte []byte) {
	c2sLogin := &protocol.C2SLogin{}
	unmarshalError := json.Unmarshal(msgByte, c2sLogin)
	if unmarshalError != nil {
		fmt.Printf("Error: unmarshal c2sLogin message error: %v", unmarshalError)
		return
	}

	fmt.Printf("Note: client request login, c2sLogin = %+v\n", c2sLogin)

	s2cLogin := &protocol.S2CLogin{
		Player:         model.PlayerData{ATK: 10, DEF: 10, HP: 10, Level: 1, MP: 10, Money: 0, RoleName: "TEST"},
		NeedCreateRole: false,
	}
	connector.Write(s2cLogin)
}

func c2sEcho(connector *net.Connector, msgByte []byte) {
	c2sEcho := &protocol.C2SEcho{}
	unmarshalError := json.Unmarshal(msgByte, c2sEcho)
	if unmarshalError != nil {
		fmt.Printf("Error: unmarshal c2sEcho message error: %v", unmarshalError)
		return
	}

	fmt.Printf("Note: client request test, c2sEcho = %+v\n", c2sEcho)

	s2cEcho := &protocol.S2CEcho{
		Value: c2sEcho.Value,
	}
	connector.Write(s2cEcho)
}

func main() {
	if pprof {
		fmt.Printf("Note: pprof is opening, please check address %v\n", pprofAddress)
		go func() {
			http.ListenAndServe(pprofAddress, nil)
		}()
	}

	go startServer()

	listener, createListenerError := net.NewListener(address, readPorcess)
	if createListenerError != nil || listener == nil {
		fmt.Printf("Error: create server listener by address %v error: %v", address, createListenerError)
		return
	}

	listener.Accept()
	listener.Close()
}
