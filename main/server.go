package main

import (
	"github.com/gorilla/websocket"
	"sync"
	"errors"
)

type Connection struct {
	wsConn *websocket.Conn
	inChan chan []byte
	outChan chan[]byte
	closeChan chan []byte

	mutex sync.Mutex
	isClosed bool
	
}

func InitConnection(wsConn *websocket.Conn)(conn *Connection,err error)  {
	conn=&Connection{
		wsConn:wsConn,
		inChan:make(chan []byte,1000),
		outChan:make(chan []byte,1000),
		closeChan:make(chan []byte,1),
	}
	//启动读协程
	go conn.readLoop()
	return 
}
//API
func (this *Connection) ReadMessage() (data []byte,err error) {
	select {
	case data = <- this.inChan:
	case <-this.closeChan:
		err = errors.New("connection is closed")
	}
	return
}
func (this *Connection) WriteMessage() (data []byte,err error)  {
	select {
	case this.outChan <- data:
	case <- this.closeChan:
		err = errors.New("connection is closed")
	}
	return

}
func (this *Connection) Close()  {
	this.wsConn.Close()
	this.mutex.Lock()
	if !this.isClosed{
		close(this.closeChan)
		this.isClosed=true
	}
	this.mutex.Unlock()

}
//内部实现

func (this *Connection) readLoop()  {
	var(
		data []byte
		err error
	)
	for{
		this.wsConn.ReadMessage()
		if _,data,err=this.wsConn.ReadMessage();err!=nil{
				goto ERR
		}
		//可能会堵塞在这里，等待inchan空闲的位置
		select {
		case this.inChan <- data:
		case <- this.closeChan:
			goto ERR
		}

	}
	ERR:
		this.Close()
}
func (this *Connection) writeMessage()  {
	var (
		data []byte
		err error
	)
	for{
		select {
		case data =<- this.outChan:
		case <- this.closeChan:
			goto ERR
		}

		if err=this.wsConn.WriteMessage(websocket.TextMessage,data);err!=nil{
			goto ERR
		}
	}
	ERR:
		this.Close()
}

