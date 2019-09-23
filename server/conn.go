package server

import (
	"errors"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

type wsConn struct {
	mutex             sync.Mutex
	connId            uint64
	wsSocket          *websocket.Conn
	inChan            chan *wsMessage
	outChan           chan *wsMessage
	closeChan         chan byte
	isClosed          bool
	lastHeartbeatTime time.Time
}

// 读循环
func (w *wsConn) readloop() {
	var (
		msgType int
		msgData []byte
		message *wsMessage
		err     error
	)
	for {
		if msgType, msgData, err = w.wsSocket.ReadMessage(); err != nil {
			w.Close()
			break
		}

		message = &wsMessage{MsgData: msgData, MsgType: msgType}
		select {
		case w.inChan <- message:
		case <-w.closeChan:
			break
		}
	}
}

// 写循环
func (w *wsConn) writeloop() {
	var (
		message *wsMessage
		err     error
	)
	for {
		select {
		case message = <-w.outChan:
			if err = w.wsSocket.WriteMessage(message.MsgType, message.MsgData); err != nil {
				w.Close()
				break
			}
		case <-w.closeChan:
			break
		}
	}
}

// 初始化
func InitWsConn(wsSocket *websocket.Conn) *wsConn {
	w := &wsConn{
		wsSocket:          wsSocket,
		inChan:            make(chan *wsMessage, 1000),
		outChan:           make(chan *wsMessage, 1000),
		closeChan:         make(chan byte),
		lastHeartbeatTime: time.Now(),
	}

	go w.readloop()
	go w.writeloop()
	return w
}

func (w *wsConn) readMsg() (msg *wsMessage, err error) {
	select {
	case msg = <-w.inChan:
	case <-w.closeChan:
		err = errors.New("Connection Loss.")
	}
	return
}

func (w *wsConn) sendMsg(msg *wsMessage) (err error) {
	select {
	case w.outChan <- msg:
	case <-w.closeChan:
		err = errors.New("Connection Loss.")
	default: // 写操作不会阻塞, 因为channel已经预留给websocket一定的缓冲空间
		err = errors.New("Connection full.")
	}
	return
}

func (w *wsConn) Close() {
	w.wsSocket.Close()
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if !w.isClosed {
		w.isClosed = true
		close(w.closeChan)
	}
}

func (w *wsConn) IsAlive() bool {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	// 连接已关闭 或者 太久没有心跳
	if w.isClosed || time.Now().Sub(w.lastHeartbeatTime) > time.Second*60 {
		return false
	}
	return true
}

func (w *wsConn) KeepAlive() {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.lastHeartbeatTime = time.Now()
}
