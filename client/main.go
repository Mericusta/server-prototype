package main

import (
	"encoding/json"
	"fmt"
	"sync"

	"server-prototype/net"
	"server-prototype/protocol"
)

// global value

var stepWaitGroup sync.WaitGroup
var clientConnectorMap sync.Map

var msgCallbackMap map[protocol.MsgID]func([]byte)

// test data

var address string = "127.0.0.1:6666"
var clientCount int = 1
var multiRequestCount int = 1

// client process

func init() {
	stepWaitGroup = sync.WaitGroup{}
	msgCallbackMap = make(map[protocol.MsgID]func([]byte))
	registCallback()
}

func registCallback() {
	msgCallbackMap[protocol.MessageEcho] = s2cEcho
	msgCallbackMap[protocol.MessageLogin] = s2cLogin
}

func readPorcess(connector *net.Connector, tag protocol.MsgID, byteDate []byte) {
	if tag == protocol.MessageOffline {
		serverKick(connector, byteDate)
	} else {
		messageCallback, hasCallback := msgCallbackMap[tag]
		if !hasCallback {
			fmt.Printf("Warn: message %v does not have callback\n", tag)
			return
		}
		messageCallback(byteDate)
	}
}

// build connector

func online(index int, connector *net.Connector) {
	c2sOnline := &protocol.C2SOnline{}
	connector.Write(c2sOnline)
}

func offline(index int, connector *net.Connector) {
	c2sOffline := &protocol.C2SOffline{}
	connector.Write(c2sOffline)
}

func serverKick(connector *net.Connector, byteData []byte) {
	s2cOffline := &protocol.S2COffline{}
	unmarshalError := json.Unmarshal(byteData, s2cOffline)
	if unmarshalError != nil {
		fmt.Printf("Error: unmarshal s2cOffline message error: %v\n", unmarshalError)
	}
	connector.Close()
}

// operation

func echo(index int, connector *net.Connector, value int) {
	c2sEcho := &protocol.C2SEcho{
		Value: value,
	}
	connector.Write(c2sEcho)
}

func login(index int, connector *net.Connector, username, password string) {
	c2sLogin := &protocol.C2SLogin{
		Username: username,
		Password: password,
	}
	connector.Write(c2sLogin)
}

// request callback

func s2cLogin(msgByte []byte) {
	s2cLogin := &protocol.S2CLogin{}
	unmarshalError := json.Unmarshal(msgByte, s2cLogin)
	if unmarshalError != nil {
		fmt.Printf("Error: unmarshal s2cLogin message error: %v", unmarshalError)
		return
	}
}

func s2cEcho(msgByte []byte) {
	s2cEcho := &protocol.S2CEcho{}
	unmarshalError := json.Unmarshal(msgByte, s2cEcho)
	if unmarshalError != nil {
		fmt.Printf("Error: unmarshal s2cEcho message error: %v", unmarshalError)
		return
	}
}

// multi client operation

func multiClientConnectServerConcurrently() {
	clientWaitGroup := sync.WaitGroup{}
	clientWaitGroup.Add(clientCount)
	for index := 0; index != clientCount; index++ {
		go func(index int) {
			defer clientWaitGroup.Done()
			fmt.Printf("Note: client %v connect server\n", index)
			connector, createConnectorError := net.NewConnector(address, readPorcess)
			if createConnectorError != nil || connector == nil {
				fmt.Printf("Error: create client %v connector by address %v error: %v\n", index, address, createConnectorError)
				return
			}
			if _, hasConnector := clientConnectorMap.LoadOrStore(index, connector); hasConnector {
				fmt.Printf("Warn: client %v already has connector\n", index)
			}
			go connector.HandleWrite()
			go connector.HandleRead()
		}(index)
	}
	clientWaitGroup.Wait()
	stepWaitGroup.Done()
}

func multiClientOnlineConcurrently() {
	clientWaitGroup := sync.WaitGroup{}
	clientWaitGroup.Add(clientCount)
	clientConnectorMap.Range(func(index, connector interface{}) bool {
		go func(index int) {
			defer clientWaitGroup.Done()
			fmt.Printf("Note: client %v is online\n", index)
			online(index, connector.(*net.Connector))
		}(index.(int))
		return true
	})
	clientWaitGroup.Wait()
	stepWaitGroup.Done()
}

func multiClientOfflineConcurrently() {
	clientWaitGroup := sync.WaitGroup{}
	clientWaitGroup.Add(clientCount)
	clientConnectorMap.Range(func(index, connector interface{}) bool {
		go func(index int) {
			defer clientWaitGroup.Done()
			fmt.Printf("Note: client %v offline\n", index)
			offline(index, connector.(*net.Connector))
		}(index.(int))
		return true
	})
	clientWaitGroup.Wait()
	stepWaitGroup.Done()
}

func multiClientRequestLoginConcurrently() {
	clientWaitGroup := sync.WaitGroup{}
	clientWaitGroup.Add(clientCount)
	clientConnectorMap.Range(func(index, connector interface{}) bool {
		go func(index int) {
			defer clientWaitGroup.Done()
			fmt.Printf("Note: client %v request login\n", index)
			login(index, connector.(*net.Connector), fmt.Sprintf("user%v", index), "test")
		}(index.(int))
		return true
	})
	clientWaitGroup.Wait()
	stepWaitGroup.Done()
}

func multiClientRequestTestConcurrently() {
	clientWaitGroup := sync.WaitGroup{}
	clientWaitGroup.Add(clientCount)
	clientConnectorMap.Range(func(index, connector interface{}) bool {
		go func(index int) {
			defer clientWaitGroup.Done()
			fmt.Printf("Note: client %v request test\n", index)
			echo(index, connector.(*net.Connector), index)
		}(index.(int))
		return true
	})
	clientWaitGroup.Wait()
	stepWaitGroup.Done()
}

func clientMultiRequestTestConcurrently(index int, connector *net.Connector) {
	timeWaitGroup := sync.WaitGroup{}
	timeWaitGroup.Add(multiRequestCount)
	for time := 0; time != multiRequestCount; time++ {
		go func(time int) {
			defer timeWaitGroup.Done()
			fmt.Printf("Note: client %v multi request test value %v\n", index, time)
			echo(index, connector, time)
		}(time)
	}
	timeWaitGroup.Wait()
}

func multiClientMultiRequestTestConcurrently() {
	clientWaitGroup := sync.WaitGroup{}
	clientWaitGroup.Add(clientCount)
	clientConnectorMap.Range(func(index, connector interface{}) bool {
		go func(index int) {
			defer clientWaitGroup.Done()
			fmt.Printf("Note: client %v multi request test concurrently\n", index)
			clientMultiRequestTestConcurrently(multiRequestCount, connector.(*net.Connector))
		}(index.(int))
		return true
	})
	clientWaitGroup.Wait()
	stepWaitGroup.Done()
}

func main() {
	stepWaitGroup.Add(5)
	multiClientConnectServerConcurrently()
	// multiClientOnlineConcurrently()
	// multiClientRequestLoginConcurrently()
	// multiClientRequestTestConcurrently()
	// multiClientOfflineConcurrently()

	stepWaitGroup.Wait()
	clientConnectorMap.Range(func(index, connector interface{}) bool {
		fmt.Printf("Note: client %v close connector\n", index)
		connector.(*net.Connector).Close()
		return true
	})
}
