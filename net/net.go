package net

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"

	"server-prototype/protocol"
	"server-prototype/utility"
)

// Connector for each client
type Connector struct {
	Address     string
	Connection  net.Conn
	close       bool
	readProcess func(*Connector, protocol.MsgID, []byte)
	SendChan    chan []byte
}

func NewConnector(address string, readProcess func(*Connector, protocol.MsgID, []byte)) (*Connector, error) {
	connection, dailError := net.Dial("tcp", address)
	if dailError != nil {
		return nil, dailError
	}
	return &Connector{Connection: connection, readProcess: readProcess, SendChan: make(chan []byte)}, nil
}

func (connector *Connector) HandleWrite() {
	for {
		select {
		case byteData, ok := <-connector.SendChan:
			if !ok {
				return
			}
			connector.write(byteData)
		}
	}
}

func (connector *Connector) HandleRead() {
	if connector.readProcess == nil {
		fmt.Printf("Error: connector read process is nil\n")
		return
	}
	for {
		if connector.close {
			break
		}
		tagByteData := make([]byte, protocol.TCPPacketDataTagSize)
		_, readTagError := connector.Connection.Read(tagByteData)
		if readTagError != nil {
			if readTagError != io.EOF {
				fmt.Printf("Error: read packet tag occurs error: %v\n", readTagError)
			}
			break
		}
		tag := binary.BigEndian.Uint32(tagByteData)

		lengthByteData := make([]byte, protocol.TCPPacketDataLengthSize)
		_, readLengthError := connector.Connection.Read(lengthByteData)
		length := binary.BigEndian.Uint32(lengthByteData)
		if readLengthError != nil {
			if readLengthError != io.EOF {
				fmt.Printf("Error: read packet length occurs error: %v\n", readLengthError)
			}
			break
		}

		valueByteData := make([]byte, int(length))
		_, readValueError := connector.Connection.Read(valueByteData)
		if readValueError != nil {
			if readValueError != io.EOF {
				fmt.Printf("Error: read packet value occurs error: %v\n", readValueError)
			}
			break
		}

		connector.readProcess(connector, protocol.MsgID(tag), valueByteData)
	}
}

func (connector *Connector) Write(message protocol.IMessage) {
	byteData, encodeError := utility.EncodeStruct(message, message.Tag())
	if encodeError != nil {
		fmt.Printf("Error: encode message %v occurs error: %v\n", message.Tag(), encodeError)
		return
	}
	connector.SendChan <- byteData
}

func (connector *Connector) write(byteData []byte) {
	connector.Connection.Write(byteData)
}

func (connector *Connector) Close() {
	close(connector.SendChan)
	connector.Connection.Close()
	connector.close = true
}

// Listener for each each server
type Listener struct {
	Address      string
	Listener     net.Listener
	readProcess  func(*Connector, protocol.MsgID, []byte)
	Count        int32
	ConnectorMap sync.Map
}

func NewListener(address string, readProcess func(*Connector, protocol.MsgID, []byte)) (*Listener, error) {
	listener, listenError := net.Listen("tcp", address)
	if listenError != nil {
		return nil, listenError
	}
	return &Listener{Address: address, Listener: listener, readProcess: readProcess}, nil
}

func (listener *Listener) Accept() {
	for {
		connection, acceptError := listener.Listener.Accept()
		if acceptError != nil {
			fmt.Printf("Error: accept connection error: %v\n", acceptError)
			continue
		}
		connector := &Connector{Address: listener.Address, Connection: connection, readProcess: listener.readProcess, SendChan: make(chan []byte)}
		go connector.HandleRead()
		go connector.HandleWrite()
		listener.ConnectorMap.Store(atomic.AddInt32(&listener.Count, 1), connector)
	}
}

func (listener *Listener) Close() {
	listener.Listener.Close()
	listener.ConnectorMap.Range(func(key interface{}, connector interface{}) bool {
		fmt.Printf("Note: Close connector %v\n", key.(int32))
		connector.(*Connector).Close()
		return true
	})
}
